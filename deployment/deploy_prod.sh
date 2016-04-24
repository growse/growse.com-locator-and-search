scp www.growse.com www-data@www.growse.com:/tmp/growse-web-go
scp static/assets.tgz www-data@www.growse.com:/tmp/growse-web-assets.tgz
ssh www-data@www.growse.com "sudo systemctl stop www-growse-com && mv /tmp/growse-web-go /var/www/growse-web/app/growse-web-go && mv /tmp/growse-web-assets.tgz /var/www/growse-web/static/ && cd /var/www/growse-web/static/ && tar -zxvf growse-web-assets.tgz && sudo systemctl start www-growse-com"
