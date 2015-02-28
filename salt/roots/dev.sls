packages:
  pkg.installed:
    - pkgs:
        - git
        - make
        - mercurial

/home/vagrant/.hushlogin:
  file.managed:
    - contents: ""
    - user: vagrant
    - group: vagrant
    - mode: 0644      

/home/vagrant/.bash_profile:
  file.managed:
    - source: salt://files/dot.bash_profile
    - user: vagrant
    - group: vagrant
    - mode: 0644

/home/vagrant/.local/bin:
  file.directory:
    - user: vagrant
    - group: vagrant
    - makedirs: true

/go:
  file.directory:
    - user: vagrant
    - group: vagrant
    - recurse:
        - user
        - group
