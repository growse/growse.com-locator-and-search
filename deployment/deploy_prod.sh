scp growse-web-go www-data@www.growse.com:/tmp/growse-web-go
ssh www-data@www.growse.com "sudo supervisorctl stop www.growse.com-golang && mv /tmp/growse-web-go /var/www/growse-web-go/growse-web-go && cd /var/www/growse-web-src/ && git pull && cd static && grunt && sudo supervisorctl start www.growse.com-golang"
