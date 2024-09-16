package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramGroupIDstr := os.Getenv("TELEGRAM_GROUP_ID")
	encryptionPass := os.Getenv("ENCRYPTION_PASS")

	if telegramBotToken == "" || telegramGroupIDstr == "" || encryptionPass == "" {
		log.Fatal("One or more environment variables are not set")
	}

	// Convert telegramGroupID from string to int64
	telegramGroupID, err := strconv.ParseInt(telegramGroupIDstr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_GROUP_ID format: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// Generate backup, encrypt it, and send it
	backupPath, err := backupSSHDirectory()
	if err != nil {
		log.Fatalf("Backup failed: %v", err)
	}
	defer os.Remove(backupPath) // Ensure backup file is removed after use

	encryptedFilePath, err := encryptWithGPG(backupPath, encryptionPass)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}
	defer os.Remove(encryptedFilePath) // Ensure encrypted file is removed after use

	err = sendBackupToTelegramGroup(bot, encryptedFilePath, telegramGroupID)
	if err != nil {
		log.Fatalf("Failed to send backup to group: %v", err)
	}
}

// backupSSHDirectory creates a backup of the .ssh directory and returns the path to the backup file.
func backupSSHDirectory() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	sshDir := filepath.Join(user.HomeDir, ".ssh")
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return "", fmt.Errorf("the .ssh directory does not exist at %s", sshDir)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(os.TempDir(), "ssh_backup_"+timestamp+".tar.gz")
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary backup file: %v", err)
	}

	backupFile.Close() // Close file to be used by tar

	err = exec.Command("tar", "-czf", backupPath, "-C", filepath.Dir(sshDir), filepath.Base(sshDir)).Run()
	if err != nil {
		return "", fmt.Errorf("failed to create archive: %v", err)
	}

	return backupPath, nil
}

// encryptWithGPG encrypts the backup file with gpg and returns the path to the encrypted file.
func encryptWithGPG(filePath string, passphrase string) (string, error) {
	encryptedFilePath := filepath.Join(os.TempDir(), filepath.Base(filePath)+".gpg")
	encryptedFile, err := os.Create(encryptedFilePath)
	if err != nil {
			return "", fmt.Errorf("failed to create encrypted file: %v", err)
	}
	encryptedFile.Close()

	cmd := exec.Command("gpg", "--batch", "--yes", "--passphrase", passphrase, "-c", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to encrypt file with gpg: %v", err)
	}

	return encryptedFilePath, nil
}

// sendBackupToTelegramGroup sends the backup file to a Telegram group.
func sendBackupToTelegramGroup(bot *tgbotapi.BotAPI, filePath string, groupID int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	doc := tgbotapi.NewDocument(groupID, tgbotapi.FileReader{
		Name:   filepath.Base(filePath),
		Reader: file,
	})

	_, err = bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send file to Telegram group: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	timeNow := time.Now().Format(time.RFC3339)

	msgText := fmt.Sprintf(
		"Backup created and sent successfully!\n\nHostname: %s\nTime: %s\nBackup Size: %d bytes",
		hostname, timeNow, fileSize,
	)

	msg := tgbotapi.NewMessage(groupID, msgText)
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send additional information to Telegram group: %v", err)
	}

	return nil
}
