# Best Practices

## Teams and Channels

The SSH CA bot user needs to have write access in all of the teams used for
granting SSH access in order for it to be able to store kssh client configs
associated with each team. Since access to a team grants SSH access to servers,
it is recommended to minimize the number of users with admin or owner
permissions in the teams. Individual users of kssh only need to be given the
read permission since they do not need to be able to edit or create files
associated with a team. 

It is also recommended to mute all notifications in the configured teams in
order to minimize the number of notifications you get. 

If you are using other bots in the same teams as the SSH CA bot (or if you wish
to have normal conversation in those teams), you can use the `CHAT_CHANNEL`
environment variable in order to configure a specific chat channel for all SSH
CA messages. 

## Network Isolation

Due to the highly sensitive nature of the SSH CA bot, it is recommended to
configure firewalls in order to block all access to the server running the CA
bot. It is not recommended to use kssh to access the server of the CA bot
itself in order to make it easier to respond to any outages. 

## Realms

There are two general approaches one can take when defining realms of servers.
The first approach (described in the getting started directions) is to define
realms for staging and production. This approach is useful for the common
scenario where all developers should be given access to the staging environment
but only certain people should be given access to production. The second
approach is a more granular approach where you can define realms associated
with teams.  For example, one could have a realm of web servers, a realm of
database servers, ... where a specific group of people is responsible for each
class of server. 
