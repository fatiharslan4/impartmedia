## Git workflow Strategy

### Master
This we will kept as development branch. Every new feature or fix branch should be base from the tip f master branch.

| Create From  | Merge Back into |
| ------------ | ------------- |
| - | - |

### Feature/Fix
Each new feature should reside in its own branch, which can be merged into the **master** branch. But, instead of branching off of master, feature branches use develop as their parent branch. When a feature is complete, it gets merged back into master.

When complete the development should create a merge request into master. Every new feature or fix should create with **branch name** and should add the **jira** ticket number. 

| Create From  | Merge Back into |
| ------------ | ------------- |
| Master | master |

Steps to follow
- Before creating a new branch, pull all the latest update from master branch
- Create a new branch from the tip of master
- If you are still working on the feature branch and new updates are merged with master, then user **Merge** or **Rebase** get fetch latest updates into the feature branch
- Once the development completed, push the codes and open an MR for review
- Once your MR is passed the pipline, should merge the code.
**No review or merge if the MR is not passed the pipline** - Pipline will be implmented soon

### ios-dev
    This branch is for ios-dev.

### staging
    Staging 

### Pre Production
This will have the code for pre production. Once the staging env is fix, pull the changes from stag.
- All the pre-production checking are done with this branch

| Create From  | Merge Back into |
| ------------ | ------------- |
| Staging | - |


### Production

Is the production branch. No direct push is not allowed here

| Create From  | Merge Back into |
| ------------ | ------------- |
| Pre Production | - |

### Hotfix
The hot-fix branch is the only branch that is created from the **Production** branch and directly merged to the Production branch instead of the develop branch. It is used only when you have to quickly patch a production issue. An advantage of this branch is, it allows you to quickly deploy a production issue without disrupting others’ workflow or without having to wait for the next release cycle.

| Create From  | Merge Back into |
| ------------ | ------------- |
| - | Production |


### Bugs workflow
What happens a bug exists in the staging,pre-production or production
- Should create a new feature branch for the bug
- Work with the fix , push and open a merge request.
- Once it was merged with master ( the development branch ) it should be merged to staging, pre-production and finally production



## Pipline Workflow

We have implemented a pipline process. When a new push comes into the branches, then it will check all teh unit tests are working fine.

work flow file **dev.workflow.yml** is the workflow for unit tests

And we have implemented another workflow action with **deployment.workflow.yml**. 
This will initiate a docker build image and push into the aws ECR. For this workflow we have to setup repository secrets from following url **<repository url>/settings/secrets/actions**

Create new variables as follow
- AWS_ACCESS_KEY_ID : access key
- AWS_SECRET_ACCESS_KEY : secret key
- AWS_ECR_REPO  : ECR repository name

Once you configure this, the pipline will create a docker image and pushed into ecr when each and every merge happens on the following branches.
- staging
- ios-dev
- pre-production
- production 