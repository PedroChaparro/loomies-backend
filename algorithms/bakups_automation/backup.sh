#!/bin/bash

# Needed parameters
HOSTS=$1
BACKUP_DIR=$2
USER=$3
DRIVE_FOLDER_ID=$4
PASSWORD=$5

# Get current date in format YYYY-MM-DD
DATE=`date +%Y-%m-%d`
echo "Creating $DATE backup ⏳"

# Execute mongo dump command
mongodump --host $HOSTS --out $BACKUP_DIR/$DATE --username $USER --password $PASSWORD
echo "$DATE Backup created ✅"

# Compress backup folder
cd $BACKUP_DIR
tar -zcvf $DATE.tar.gz $DATE
echo "$DATE Backup compressed 📥"

# Run gdrive upload command
gdrive files upload $BACKUP_DIR/$DATE.tar.gz --parent $DRIVE_FOLDER_ID
echo "$DATE Backup uploaded to Google Drive ☁️"