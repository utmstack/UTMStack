from abc import ABC, abstractmethod

from util.misc import get_config, get_module_group


class Integration(ABC):
    """
    The Product interface declares the operations that all concrete products
    must implement.
    """

    @abstractmethod
    def get_integration_config(self):
        pass

    def get_input_integration(self, integration, principal_key, group):
        agent_cfg = get_config(integration, group)
        if not agent_cfg.get(principal_key) or agent_cfg.get(principal_key) == "":
            return None
        return agent_cfg

    def get_input_group(self, integration):
        return get_module_group(integration)
