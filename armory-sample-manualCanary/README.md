#cdCon-CD-as-a-Service-demo
Get started with using CD-as-a-Service by deploying a sample application to your cluster.

## Without Using Github CLI
1. Setup RNA using the docs here https://docs.armory.io/cd-as-a-service/setup/get-started/#prepare-your-deployment-target
2. Create a fork of this repository from Github UI. 
3. Create new machine to machine credentials
Go to https://console.cloud.armory.io/configuration > Client Credentials
4. You can select the preconfigured scope group `Deployments using Spinnaker` or manually select the following:
```
manage:deploy
read:infra:data
exec:infra:op
read:artifacts:data
```
4. Copy the Client ID and Client Secret so that they can be saved as secrets later. 
5. Navigate to the Github page of your fork.
`https://github.com/<github-user-name>/cdCon-cdaas-demo`
6. Navigate to the Repository settings → Secrets → Actions
7. Create 2 new secrets
   1. `CDAAS_CREDENTIAL_ID` and paste the value for Client ID created in Step 3 
   2. `CDAAS_CREDENTIAL_SECRET` and paste the value for Client Secret created in Step 3
8. Open `deploy.yml` and edit the `account` field to the name of the RNA. This can be found at https://console.cloud.armory.io/configuration > Agents (in the left side panel)
   `account: <your RNA name>`
9. Commit and push the changes to the repository in the main branch. 
10. Check the Action tab for the ongoing action that will deploy the sample application to your Kubernetes cluster.
    
## Using GitHub (gh) CLI
    
1. Setup RNA using the docs here https://docs.armory.io/cd-as-a-service/setup/get-started/#prepare-your-deployment-target
2. Create a fork of this repo
    `gh repo fork https://github.com/armory/cdCon-cdaas-demo --fork-name cdCon-cdaas-demo --clone`
3. Create new machine to machine credentials 
   1. Go to https://console.cloud.armory.io/configuration > Client Credentials 
   2. You can select the preconfigured scope group Deployments using Spinnaker or manually select the following:
      ```
      manage:deploy
      read:infra:data
      exec:infra:op
      read:artifacts:data
      ```
   3. Copy the Client ID and Client Secret so that they can be saved as secrets. 
4. Navigate into the directory of the forked repo. 
5. Create github secrets for Client ID and Client Secret generated in Step 3
   1. `gh secret set CDAAS_CREDENTIAL_ID -a actions -b <Client_ID>`
   2. `gh secret set CDAAS_CREDENTIAL_SECRET -a actions -b <Client_Secret>`
6. Open `deploy.yml` and edit the `account` field to the name of the RNA. This can be found at https://console.cloud.armory.io/configuration > Agents (in the left side panel)
   `account: <your RNA name>`
7. Commit and push the changes to the repository in the main branch.
8. Check the Action tab for the ongoing action that will deploy the sample application to your Kubernetes cluster.
