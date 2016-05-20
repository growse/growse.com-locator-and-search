#!/usr/bin/env bash
scp www.growse.com www-data@www.growse.com:/tmp/growse-web-go
scp static/assets.tgz www-data@www.growse.com:/tmp/growse-web-assets.tgz
scp templates/templates.tgz www-data@www.growse.com:/tmp/growse-web-templates.tgz
scp databasemigrations/databasemigrations.tgz www-data@www.growse.com:/tmp/growse-web-databasemigrations.tgz
ssh www-data@www.growse.com "mv /tmp/growse-web-databasemigrations.tgz /var/www/growse-web/databasemigrations/ && cd /var/www/growse-web/databasemigrations/ && tar -zxvf growse-web-databasemigrations.tgz && sudo systemctl stop www-growse-com && mv /tmp/growse-web-go /var/www/growse-web/app/growse-web-go && sudo systemctl start www-growse-com && mv /tmp/growse-web-assets.tgz /var/www/growse-web/static/ && cd /var/www/growse-web/static/ && tar -zxvf growse-web-assets.tgz && mv /tmp/growse-web-templates.tgz /var/www/growse-web/templates/ && cd /var/www/growse-web/templates/ && tar -zxvf growse-web-templates.tgz"
