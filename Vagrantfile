# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "chef/ubuntu-14.04"
  config.vm.network :forwarded_port, guest: 8080, host: 8080
  config.vm.synced_folder ".", "/go/src/github.com/martinp/ifconfig"
  config.vm.synced_folder "salt/roots/", "/srv/salt/"
  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", "512"]
  end
  config.vm.provision :salt do |salt|
    salt.minion_config = "salt/minion.yml"
    salt.run_highstate = true
    salt.colorize = true
  end
end
