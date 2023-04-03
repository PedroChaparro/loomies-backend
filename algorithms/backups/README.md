## 💾 Backups

This folder contains the bash script used to backup the database.

- **Execution period:** At 01:00 AM, only on Tuesday, Thursday, and Saturday
- **Required arguments:**

| Pos. | Name             | Description                                                     | Example                                               |
| ---- | ---------------- | --------------------------------------------------------------- | ----------------------------------------------------- |
| 1    | Hosts            | Comma separated mongodb hosts                                   | mongodb0.example.com:27017,mongodb1.example.com:27017 |
| 2    | Backup Directory | Directory to save the backup output                             | /home/server/backups                                  |
| 3    | User             | Database user with permissions to execute the mongodump command | admin                                                 |
| 4    | Drive Folder Id  | Google Drive folder id to save the backup                       | 1a2b3...                                              |
| 5    | Password         | Database user's password                                        | password                                              |
