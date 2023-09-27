from abc import ABC, abstractmethod


class Creator(ABC):
    """
    The Creator class declares the factory method that is supposed to return an
    object of a Product class. The Creator's subclasses usually provide the
    implementation of this method.
    """

    @abstractmethod
    def create_integration(self, integration):
        """
        Note that the Creator may also provide some default implementation of
        the factory method.
        """
        pass
