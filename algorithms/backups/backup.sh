#!/bin/bash

# Needed parameters (passed as cli arguments)
HOSTS=$1
BACKUP_DIR=$2
USER=$3
DRIVE_FOLDER_ID=$4
PASSWORD=$5

# Get current date in format YYYY-MM-DD:HH:MM:SS
DATE=$(date +"%Y-%m-%d:%H:%M:%S")
echo "Creating backup ⏳"

# Execute mongo dump command
mongodump --host $HOSTS --out $BACKUP_DIR/$DATE --username $USER --password $PASSWORD
STATUS=$?

if [ $STATUS -ne 0 ]; then
  echo "🚨 Error creating backup"
  exit 1
fi

echo "Backup created ✅"

# Compress backup folder
cd $BACKUP_DIR
tar -zcvf $DATE.tar.gz $DATE
STATUS=$?

if [ $STATUS -ne 0 ]; then
  echo "🚨 Error compressing backup"
  exit 1
fi

echo "Backup compressed 📥"

# Run gdrive upload command
gdrive files upload $BACKUP_DIR/$DATE.tar.gz --parent $DRIVE_FOLDER_ID
STATUS=$?

if [ $STATUS -ne 0 ]; then
  echo "🚨 Error uploading backup to Google Drive"
  exit 1
fi

echo "Backup uploaded to Google Drive ☁️"