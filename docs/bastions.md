# Jumpboxes and Bastion Hosts

kssh should work correctly with jumpboxes and bastion hosts as long as they are configured to trust the SSH CA and the usernames are correct. For example:

```
kssh -J developer@jumpbox.example.com developer@server.internal
```

This can also be made easier by setting the kssh default ssh-username locally, then you won't have to specify it for each server. 

```
kssh --set-default-user developer
kssh -J jumpbox.example.com server.internal
```