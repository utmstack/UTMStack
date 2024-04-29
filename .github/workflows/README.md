# New Workflows

### Branches involved
***main***<br>
***feature/\*\*/\*\****<br>
***hotfix/\*\*/\*\****<br>

## Feature:
__Note__: only one feature per version.
1.  Create branch from main for the feature.
2. Every developer related to the feature works on this branch. 
3. When they finish they push.
4. This push activates the DEV actions, generates the DEV images, and is automatically deployed to DEV environment.
5. When everything is tested, PR is made to main.
6. After this PR is approved, the QA actions are activated, and the QA images are generated.
7. QA is updated manually to see that everything works well.
11. If everything works well, merge to main activate RC actions and generate RC images.
12. After test on RC, create a new release that will activate release actions.

## Hotfix
1. Create branch from main for the hotfix.
2. Every developer related to the hotfix works on this branch. 
3. When they finish, they make a PR to main.
4. After this PR is merged to main, the rc actions are activated, and the rc images are generated.
5. RC is updated manually to see that everything works well and we test it in RC.
6. If everything works well, create a new release that will activate release actions.


### Actions
dev->on push ***feature/\*\*/\*\**** branch<br>
qa->on pull-request: approved main branch<br>
rc->on push: main branch<br>
release-> on release<br>