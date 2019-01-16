![Vaulter](https://github.com/hashicorp/vault/blob/f22d202cde2018f9455dec755118a9b84586e082/Vault_PrimaryLogo_Black.png)

# Vaulter
> friendliest console client for vault

Vaulter extends default Hashicorp Vault client, implements additional methods like list all backend path, dynamically read value, search in all backend and so on.


## Features

Okey, whats new? Why can i use default vault client? 
* Despicable list subcommand 
* Not useful
* Really not useful (as mentioned above)

## Installing / Getting started

Installation looks like building module - clone repo, than run
 `go build vaulter.go`
If you haven't Golang environment, download precompiled packages.  

Also you may have default vault environment variables - `VAULT_ADDR` and `VAULT_TOKEN`.  

### Deploying / Publishing

Copy in to your $PATH.

```
for os x:  /usr/local/bin
for linux: 
for windows: 
```

Than use like regular binary file.

## Main features 

### vaulter verify

Verify your connection to vault. Make this with first launch. 

Example:
```bash
$ vaulter verify

Lookup url: http://yuor_path_here:8200/v1/auth/token/lookup-self
Request id: f1452539-fe68-3537-7d38-327932d47b3f
Request token: your_token_here
Policies: root 
```

### vaulter tree [backend_name]
Create tree of keys in your backend. It will show only keys, without subfolders.   
Specify your backend name as parameter. 

```bash
$ vaulter tree backend
backend/vultr/one
backend/wordpress/blog
backend/devops/omapi/m1
backend/devops/ovh
backend/devops/mikrotik
...
```
Tree parameter can me more sensitive: 
```bash
$ vaulter tree backend/devops
backend/devops/omapi/m1
backend/devops/ovh
backend/devops/mikrotik
...
```
Tree output you can grep, like 
```  
$ vaulter tree backend | grep devops
backend/devops/omapi/m1
backend/devops/ovh
backend/devops/mikrotik
```

### vaulter search [backend_name] [regex]

It searches for pattern in your backend and show all key-values in path.  
This can be used if you want to make speedily search, but cant remember path value. 

Example:
```bash
$ vaulter search backend .*mikrotik.*

backend/devops/mikrotik
password: pass
ip: 192.168.1.1:8080
login: admin
```

Or more complex regexp: 
```bash
$ vaulter search backend omapi.*m1

backend/devops/omapi/m1
key: key==
ip: 10.10.17.1
```
