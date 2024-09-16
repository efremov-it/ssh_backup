### Simple .ssh directory backup

1. copy binary to /home/$USER/.ssh/ direcory

wget https://github.com/efremov-it/ssh_backup/releases/download/v0.0.2/ssh_backup -O /home/$USER/.ssh/ssh_backup
chmod u+x /home/$USER/.ssh/ssh_backup

2. add .env file to /home/$USER/.ssh/

wget https://github.com/efremov-it/ssh_backup/raw/master/.env.example -O /home/$USER/.ssh/.env

3. Change values

TELEGRAM_BOT_TOKEN=bot_token
TELEGRAM_GROUP_ID=group_id
ENCRYPTION_PASS=secret_password

3. Create cron job.

`crontab -e`

0 23 * * * /home/$(whoami)/.ssh/ssh_backup -env /home/$(whoami)/.ssh/.env


### Dectrypt tar file

gpg --batch --yes --passphrase "secret_password" -d -o decrypted_file_name original_encrypted_file.gpg

