#!/bin/bash

# Needed parameters
HOSTS=$1
BACKUP_DIR=$2
USER=$3
PASSWORD=$4
DATABASE=$5

# Get current date in format YYYY-MM-DD
DATE=`date +%Y-%m-%d`
echo "Creating $DATE backup ⏳"

# Execute mongo dump command
mongodump --host=$HOSTS --db $DATABASE --username $USER --password $PASSWORD --out $BACKUP_DIR/$DATE
echo "$DATE Backup created ✅"

# Compress backup folder
tar -zcvf $BACKUP_DIR/$DATE.tar.gz $BACKUP_DIR/$DATE
echo "$DATE Backup compressed 📥"

# Run gdrive upload command
gdrive upload $BACKUP_DIR/$DATE.tar.gz
echo "$DATE Backup uploaded to Google Drive ☁️"