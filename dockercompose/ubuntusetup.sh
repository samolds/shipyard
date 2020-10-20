#!/bin/sh


# This script is to be used on a fresh ubuntu install to download docker
# and docker-compose


# from https://docs.docker.com/engine/install/ubuntu
# To install docker
# update apt
sudo apt-get update

sudo apt-get -y install \
  apt-transport-https \ ca-certificates \
  curl \
  gnupg-agent \
  make \
  software-properties-common

# add Docker's official gpg key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

# verify last 8 characters of gpg key
#sudo apt-key fingerprint xxxxxxxx

# set up the stable repo
sudo add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) \
  stable"

# install the latest version of docker engine
sudo apt-get update
sudo apt-get -y install docker-ce
# sudo apt-get -y install docker-ce # docker-ce-cli containerd.io

# to use docker as a non-root user, add the current user to the docker group
sudo usermod -aG docker ${USER}

# https://docs.docker.com/compose/install
# To install docker-compose
sudo curl -L \
  "https://github.com/docker/compose/releases/download/1.26.2/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose

sudo chmod +x /usr/local/bin/docker-compose

mkdir ~/shipyard
