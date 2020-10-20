# Deployment Steps

The production build and deployment process was developed with the help of
[this TestDriven.io series](https://testdriven.io/blog/django-docker-https-aws).
Look there for more details about each of the steps listed below.


### Steps:
1. Sign in to the [AWS EC2 Console](https://console.aws.amazon.com/ec2) and set
   up a new instance.
   - use an LTS Ubuntu image
   - add a new security group
   - add rule for `HTTP` -> "Source: Anywhere"
   - add rule for `HTTPS` -> "Source: Anywhere"
   - create a new key pair
2. Configure your fresh instance.
   - `AWS_IP=the.aws.box.ip`
   - `chmod 400 newkeypair.pem`
   - `scp -i newkeypair.pem ubuntusetup.sh ubuntu@${AWS_IP}:~/.`
   - `ssh -i newkeypair.pem ubuntu@${AWS_IP}`
   - `./ubuntusetup.sh`
3. Copy src and build
   - `scp -i newkeypair.pem \
       -r ./{.env,Makefile,docker-compose.*,nginx,src} \
       ubuntu@${AWS_IP}:~/shipyard`
4. Update env variables in `ubuntu@${AWS_IP}:~/shipyard/.env`
5. Update Auth0 client whitelist for associated Auth0 authentication provider
   to include the new AWS host
6. Set up ElasticIP ...
7. Set up IAM Roles and ECR as the Docker image repo
8. Add DNS records
9. Set up RDS to use instead of containerized Postgres
10. Set up security group
11. Test LetsEncrypt Cert generation using a staging cert
12. Update docker-compose.prod.yml to pull docker images from ECR and remove
    the db image
13. Update necessary .env/ file to connect to RDS db, to verify LetsEncrypt and
    Virtual Hosts, and to include a proxy default email
14. Build and push the containers
15. Pull the images on the production box
16. Start the whole enchilade with:
    `docker-compose -f docker-compose.prod.yml up -d`
