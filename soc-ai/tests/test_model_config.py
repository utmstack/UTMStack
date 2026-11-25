import json

from socai.crypto import Cipher
from socai.model_config import ModelConfigs

C = Cipher("the-key")


def doc(**kw):
    return json.dumps(kw)


def test_a_tenant_without_its_own_uses_the_instance_default():
    cfgs = ModelConfigs.parse(doc(default={"provider": "openai", "model": "gpt", "url": "https://x/v1/chat/completions"}), C)

    assert cfgs.for_tenant("any-tenant").model == "gpt" and cfgs.for_tenant(None).model == "gpt"


def test_a_tenant_with_its_own_uses_it_and_the_others_keep_the_default():
    cfgs = ModelConfigs.parse(doc(
        default={"provider": "openai", "model": "gpt", "url": "u"},
        tenants={"t1": {"provider": "anthropic", "model": "claude", "url": "v"}}), C)

    assert cfgs.for_tenant("t1").model == "claude" and cfgs.for_tenant("t2").model == "gpt"


def test_secrets_are_decrypted_and_the_rest_is_left_alone():
    cfgs = ModelConfigs.parse(doc(default={
        "provider": "threatwinds", "model": "m", "url": "u", "api_key": C.encrypt("sk-1"), "auth_type": "none",
        "custom_headers": {"id": C.encrypt("inst-7"), "key": C.encrypt("inst-secret"), "topk": "50"}}), C)
    d = cfgs.for_tenant("t")

    assert d.api_key == "sk-1" and d.auth_type == "none"
    assert d.headers == {"id": "inst-7", "key": "inst-secret", "topk": "50"}, "a plain header value is not mangled"


def test_the_limits_have_the_defaults_the_plugin_had():
    d = ModelConfigs.parse(doc(default={"provider": "openai", "model": "m", "url": "u"}), C).for_tenant("t")

    assert d.max_tokens == 4096 and d.max_steps == 12 and d.auth_type == "bearer"
    assert not d.auto_analyze and not d.allow_incidents


def test_the_settings_of_a_tenant_come_through():
    d = ModelConfigs.parse(doc(default={"provider": "openai", "model": "m", "url": "u", "max_tokens": 900,
                                        "max_tool_iterations": 5, "auto_analyze": True, "allow_incidents": True}), C).for_tenant("t")

    assert (d.max_tokens, d.max_steps, d.auto_analyze, d.allow_incidents) == (900, 5, True, True)


def test_a_provider_with_a_known_address_does_not_need_one():
    d = ModelConfigs.parse(doc(default={"provider": "openai", "model": "m"}), C).for_tenant("t")

    assert d.url == "https://api.openai.com/v1/chat/completions"


def test_a_configuration_missing_what_it_needs_is_not_a_model():
    for default in ({"provider": "", "model": "m", "url": "u"}, {"provider": "azure", "model": "m"},
                    {"provider": "openai", "model": "", "url": "u"}):
        assert ModelConfigs.parse(doc(default=default), C).for_tenant("t") is None


def test_an_own_configuration_that_is_incomplete_does_not_hide_a_good_default():
    cfgs = ModelConfigs.parse(doc(default={"provider": "openai", "model": "gpt", "url": "u"},
                                  tenants={"t1": {"provider": "openai", "model": ""}}), C)

    # The tenant's own is what it runs under, and it is not usable: that is a configuration problem to
    # show, not something to paper over with another tenant's model.
    assert cfgs.for_tenant("t1") is None
