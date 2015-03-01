ssh www-data@www.growse.com "sudo supervisorctl stop www.growse.com-golang"
scp src/growse-web-go/growse-web-go www-data@www.growse.com:/var/www/growse-web-go
ssh www-data@www.growse.com "cd /var/www/growse-web-src/ && git pull && cd static && grunt && sudo supervisorctl start www.growse.com-golang"
