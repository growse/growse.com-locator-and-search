from fabric.api import sudo, cd, run, env, local, prefix

env.hosts = ['www.growse.com']
env.use_ssh_config = True


def deploy():
    local('git rebase master deploy')
    local('git push')
    local('git checkout master')
    with cd('/home/growse/django-sites/www.growse.com'):
        with prefix('source bin/activate'):
            run('git pull')
            run('pip install --upgrade -r requirements.txt')
            run('./manage.py migrate')
            run('./manage.py collectstatic --noinput')
            sudo('supervisorctl restart www.growse.com-uwsgi')
