# New Workflows

### Branches involved
***main***<br>
***feature/\*\*/\*\****<br>
***hotfix/\*\*/\*\****<br>

## Feature:
__Note__: only one feature per version
1.  Create branch from main for the feature
2. Every developer related to the feature works on this branch. 
3. When they finish they push
4. This push activates the dev actions, generates the dev images, and is automatically deployed to dev environment
5. When everything is tested, PR is made to main
6. After this PR is approved, the rc actions are activated, and the rc images are generated
7. RC is updated manually to see that everything works well and we test it in RC
11. If everything works well, merge to main and create a tag and said tag activates release actions

## Hotfix
1. Create branch from main for the hotfix
2. Every developer related to the hotfix works on this branch. 
3. When they finish , they make a PR to main
4. After this PR is approved, the rc actions are activated, and the rc images are generated
5. RC is updated manually to see that everything works well and we test it in RC
6. If everything works well, merge to main and create a tag and said tag activates release actions


### Actions
dev->on push ***feature/\*\*/\*\**** branch<br>
rc->on pull-request: approved main branch<br>
release-> on tag v*<br>