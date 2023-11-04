"""Azure Event Hub"""


class AzureEventHub:

    def __init__(self, event_hub_connection,
                 storage_connection,
                 consumer_group,
                 storage_container):
        self.event_hub_connection = event_hub_connection
        self.storage_connection = storage_connection
        self.consumer_group = consumer_group
        self.storage_container = storage_container
