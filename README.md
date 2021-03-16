# backend
Combined backend API for ImpartWealth


### Connecting to AWS RDS DB via multi-hop SSH tunnel
The network is separated intentionally - you will only have permissions to the
bastion host.  However, the bastion host has no connectivity to the RDS Database,
so you need tunnel through an EC2 instance (one of the container hosts) first.

So, if you were to setup your ~/.ssh/config below where `10.0.148.230` is one of
the ec2 IP's in the `backend` _private_ subnet.

``` 
Host bastion.impartwealth.com
  ForwardAgent yes

Host ecs_ssh_shortcut
    HostName 10.0.148.230
    User ec2-user
    Port 22
    ForwardAgent yes
    ProxyCommand ssh -A ubuntu@bastion.impartwealth.com -W %h:%p
```

Then, Run the following command 
```bash
ssh  -L 3306:impart-dev-mysql.cluster-cnto08jmowe9.us-east-2.rds.amazonaws.com:3306 ecs_ssh_shortcut
```

That would allow you to address the remote, private DB in port 3306 
via localhost/127.0.0.1