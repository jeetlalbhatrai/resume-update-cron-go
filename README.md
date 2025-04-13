# resume-update-cron-go
for naukri login and update resumes


docker build -t resume-updater .
docker run --rm -it resume-updater
docker run --rm -it --env-file .env resume-updater



docker build --platform linux/arm64 -t resumeupdater .
docker build --platform linux/amd64 -t resumeupdater .


