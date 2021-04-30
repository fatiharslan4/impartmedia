## Git wirkflow Strategy

### Master
This we will kept as development branch. Every new feature or fix branch should be base from the tip f master branch.

| Create From  | Merge Back into |
| ------------ | ------------- |
| - | - |

### Feature/Fix
Each new feature should reside in its own branch, which can be pushed to the **master** branch. But, instead of branching off of master, feature branches use develop as their parent branch. When a feature is complete, it gets merged back into master.

When complete the development should create a merge request into master. Every new feature or fix should create with **branch name** and should add the **jira** tickete number. 

| Create From  | Merge Back into |
| ------------ | ------------- |
| Master | master |

Steps to follow
- Before creating a new branch, pull all the latest update from master branch
- Create a new branch from the tip of master
- If you are still working on the feature branch and new updates are merged with master, then user *Merge* or *Rebase* get fetch latest updates into the feature branch
- Once the development completed, push the codes and open an MR for review
- Once your MR is passed the pipline, should merge the code.
*No review or merge if the MR is not passed the pipline* - Pipline will be implmented soon

### ios-dev
    This branch is for ios-dev.

### staging
    Staging 

### Pre Production
| Create From  | Merge Back into |
| ------------ | ------------- |
| Staging | - |


### Production

Is the production branch.

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