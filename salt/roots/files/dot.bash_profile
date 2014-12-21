test -d $HOME/.local/bin || mkdir -p $HOME/.local/bin
export PATH=/vagrant/bin:$HOME/.local/bin:$PATH
export GOPATH=/vagrant
cd /vagrant
