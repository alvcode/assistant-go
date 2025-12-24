Notes and reminders system

frontend: <a href="https://github.com/alvcode/assistant-front" target="_blank">https://github.com/alvcode/assistant-front</a>

<h2>For local development</h2> 

- copy .env.example to .env

- run commands
```
// launch when deploying a project for the first time
make install

// for development
make start

// migrations
make m

// for stop development
make stop
```

<h2>For Production</h2>

- copy .env.example to .env

- run command
```
make deploy
```

- Nginx setting
```
server {

        server_name api.<domain>.<com>;

        # You need to sync this parameter with .env FILE_UPLOAD_MAX_SIZE and DRIVE_UPLOAD_MAX_SIZE
        client_max_body_size 20m;
        
        #add_header X-Frame-Options "SAMEORIGIN";
        #add_header X-Content-Type-Options "nosniff";

        location / {
                proxy_pass http://127.0.0.1:8075;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
        }
}
```

- set crontab
```
#Makes daily backups to the directory specified in .env
10 1 * * * cd /path/to/project && make backup-db

#Deletes old backups, leaving the last 5
40 5 * * * cd /path/to/project && make db-remove-old-backups

#Cleans up stale records in the database
0 4 25 * * cd /path/to/project && make cli-clean-db-p
```

