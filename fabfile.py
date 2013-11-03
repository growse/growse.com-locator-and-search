from fabric.api import sudo, cd, run, env, local, prefix

env.hosts = ['109.123.84.240']
env.user = 'growse'


def deploy():
    local('git rebase master deploy')
    local('git push')
    local('git checkout master')
    with cd('/home/growse/django-sites/www.growse.com'):
        with prefix('source bin/activate'):
            run('git pull')
            run('pip install --upgrade -r requirements.txt')
            sudo('touch /etc/uwsgi/apps-enabled/growse.com.ini')
