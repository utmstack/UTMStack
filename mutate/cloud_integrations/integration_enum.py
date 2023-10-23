"""Integration enum"""

from strenum import UppercaseStrEnum


class IntegrationEnum(UppercaseStrEnum):
    """Enumerator of integrations"""
    AZURE = 'AZURE',
    GOOGLE = 'GOOGLE',
    AWS = 'AWS',
