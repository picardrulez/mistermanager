# misterManager
golang application for pullinga golang application git repo via githook, building the applicaation, and starting it under supervisord


## installation setup
* create a user 'mistermanager' with home directory in /var/lib/mistermanager
* give user an ssh key pair
* put user's public key in github repo
* have github hit url in a hook:  my.ip.add.ress:8080/build?user=username&repo=reponame
