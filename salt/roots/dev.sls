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