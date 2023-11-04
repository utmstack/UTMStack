from cloud_integrations.azure_integration import AzureIntegration
from cloud_integrations.creator import Creator
from cloud_integrations.google_integration import GoogleIntegration
from cloud_integrations.integration import Integration
from cloud_integrations.integration_enum import IntegrationEnum


class IntegrationCreator(Creator):
    """
    Note that the signature of the method still uses the abstract product type,
    even though the concrete product is currently returned from the method. This
    way the Creator can stay independent  of concrete product classes.
    """

    def create_integration(self, integration: IntegrationEnum) -> Integration:
        integrations = {
            IntegrationEnum.AZURE: AzureIntegration,
            IntegrationEnum.GOOGLE: GoogleIntegration,
        }
        return integrations[integration]()
