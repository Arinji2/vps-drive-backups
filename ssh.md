# STEPS TO SETUP SSH PERMISSIONS

## Constants:

1. Username: backups
2. Working directory: /srv/data

## Steps:

1.  Create user \
    `sudo useradd backups`

1.  Set a password for the user \
    `sudo passwd backups`

1.  Give access to the directories you want to backup \
    `sudo chgrp backups /srv/data`

1.  Make backups group the owner group \
    `sudo chgrp backups /srv/data`

1.  Give backups read and write perms to the directory \
    `sudo chmod g+rw /srv/data`

1.  Make new files inside /srv/data copy group \
    `sudo chmod g+s /srv/data`

1.  Make new files inside /srv/data copy perms (you need to have acl installed) \
    `sudo setfacl -d -m g:backups:rwx /srv/data`
