docker.io:
  pkg:
    - installed

vagrant:
  user:
    - present
    - groups:
        - docker
    - remove_groups: False
