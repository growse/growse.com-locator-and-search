scp www.growse.com www-data@www.growse.com:/tmp/growse-web-go
scp assets.tgz www-data@www.growse.com:/tmp/growse-web-assets.tgz
ssh www-data@www.growse.com "sudo supervisorctl stop www.growse.com-golang && mv /tmp/growse-web-go /var/www/growse-web-go/growse-web-go && mv /tmp/growse-web-assets.tgz /var/www/growse-web-go && cd /var/www/growse-web-go/ && tar -zxvf growse-web-assets.tgz && sudo supervisorctl start www.growse.com-golang"
