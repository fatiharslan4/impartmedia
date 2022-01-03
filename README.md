

#

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

### Adding migrations

Migrations use https://github.com/golang-migrate/migrate 

adding a migration uses the `migrate` tool after it's installed
```bash
migrate create -ext sql -dir schemas/migrations {name}
```

will create 2 files, a `{timestamp}_{name}_up.sql` and a `{timestamp}_{name}_down.sql` - see
[best practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) for more information.

Running `script/run_local_db.sh` will tear down the current local DB,
start a new one, apply all `up` migrations, apply all `down` migrations, then
re-apply all `up` migrations - this ensures all things are ordered and re-runnable.
This script leaves the DB running.
