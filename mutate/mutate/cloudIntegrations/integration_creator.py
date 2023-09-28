from mutate.cloudIntegrations.azure_integration import AzureIntegration
from mutate.cloudIntegrations.creator import Creator
from mutate.cloudIntegrations.google_integration import GoogleIntegration
from mutate.cloudIntegrations.integration import Integration
from mutate.cloudIntegrations.integration_enum import IntegrationEnum


class IntegrationCreator(Creator):
    """
    Note that the signature of the method still uses the abstract product type,
    even though the concrete product is actually returned from the method. This
    way the Creator can stay independent  of concrete product classes.
    """

    def create_integration(self, integration: IntegrationEnum) -> Integration:
        integrations = {
            IntegrationEnum.AZURE: AzureIntegration,
            IntegrationEnum.GOOGLE: GoogleIntegration,
        }
        return integrations[integration]()
