#!/usr/bin/env bash
scp www.growse.com www-data@www.growse.com:/tmp/growse-web-go
scp templates/templates.tgz www-data@www.growse.com:/tmp/growse-web-templates.tgz
scp databasemigrations/databasemigrations.tgz www-data@www.growse.com:/tmp/growse-web-databasemigrations.tgz
ssh www-data@www.growse.com "mv /tmp/growse-web-databasemigrations.tgz /var/www/growse-web/databasemigrations/ && cd /var/www/growse-web/databasemigrations/ && tar -zxvf growse-web-databasemigrations.tgz && sudo systemctl stop www-growse-com && mv /tmp/growse-web-go /var/www/growse-web/app/growse-web-go && sudo systemctl start www-growse-com && cd /var/www/growse-web/static/ && mv /tmp/growse-web-templates.tgz /var/www/growse-web/templates/ && cd /var/www/growse-web/templates/ && tar -zxvf growse-web-templates.tgz"
