# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
  # Ensure noninteractive apt-get
  export DEBIAN_FRONTEND=noninteractive

  # Set time zone
  echo "Europe/Oslo" > /etc/timezone
  dpkg-reconfigure tzdata

  # Install packages
  apt-get -y --quiet update
  apt-get -y --quiet install git make

  # Install golang
  test -d /usr/local/go || \
      curl https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz | \
      tar -xzC /usr/local
  test -s /etc/profile.d/golang.sh || \
      echo 'export PATH=/usr/local/go/bin:$PATH' > /etc/profile.d/golang.sh
SCRIPT

Vagrant.configure("2") do |config|
  config.vm.box = "raring64-current"
  config.vm.box_url = "http://cloud-images.ubuntu.com/vagrant/raring/current/raring-server-cloudimg-amd64-vagrant-disk1.box"
  config.vm.network :forwarded_port, guest: 8080, host: 5000
  config.ssh.forward_agent = true
  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", "1024"]
  end
  config.vm.provision :shell, :inline => $script
end
