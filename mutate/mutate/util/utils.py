# pylint: disable=R1732
"""UTMStack utilities."""

import argparse
import json
import logging.handlers
import os
import time
from threading import Thread
from typing import Any, Callable, Dict, Generator, Optional

# pylama:ignore=W0611
from mutate.util.postgres import Postgres


def get_module_group(module: str):
    """Get groups of configuration module"""
    query = """select distinct group_name from
    utm_server_configurations WHERE module_name=%s;"""
    return [group[0] for group in Postgres().fetchall(query,
                                                      ( module))]


def get_config(module: str, group: str) -> Dict[str, Any]:
    """Get configuration for the given module."""
    query = """SELECT conf_short, conf_value
        FROM utm_server_configurations
        WHERE module_name=%s AND group_name=%s;"""
    configs = Postgres().fetchall(
        query, (module, group))
    cfg = {}
    for key, value in configs:
        try:
            val = json.loads(value)
        except Exception:
            val = value
        cfg[key] = val
    return cfg


def get_pipelines():
    try:
        query_result = Postgres().fetchall("""
        SELECT 
          ulp.pipeline_id,
          json_agg(json_build_object('input_plugin', uli.input_plugin, 'conf', confs))
           AS inputs,
          json_agg(DISTINCT ulf.logstash_filter) AS filters
        FROM 
          utm_logstash_pipeline AS ulp
        LEFT JOIN 
          utm_logstash_input AS uli ON ulp.id = uli.pipeline_id
        LEFT JOIN 
          (SELECT 
             ulic.input_id, 
             json_build_object(ulic.conf_key, ulic.conf_value) AS confs
           FROM 
             utm_logstash_input_configuration AS ulic
           GROUP BY 
             ulic.input_id, ulic.conf_key, ulic.conf_value
          ) AS confs ON uli.id = confs.input_id
        LEFT JOIN 
          utm_group_logstash_pipeline_filters AS ulg ON ulg.pipeline_id = ulp.id
        LEFT JOIN 
          utm_logstash_filter AS ulf ON ulf.id = ulg.filter_id
        GROUP BY 
          ulp.pipeline_id;
        """)
        return {row['pipeline_id']: dict(row) for row in query_result}

    except Exception as e:
        logging.error(f"Unexpected error occurred when trying to get pipelines: {e}")
        return {}


def get_active_pipelines():
    try:
        query_result = Postgres().fetchall("""
        SELECT ulp.id, ulp.pipeline_id
        FROM utm_logstash_pipeline AS ulp
        LEFT JOIN utm_module AS um ON ulp.module_name = um.module_name
        WHERE um.module_active IS NULL OR um.module_active = TRUE;
    """)
        return [row['pipeline_id'] for row in query_result]

    except Exception as e:
        logging.error(f"Unexpected error occurred when trying to get active pipelines: {e}")
        return []


def compare_dicts_in_unordered_lists(list1, list2):
    if len(list1) != len(list2):
        return False
    for dictionary in list1:
        if dictionary not in list2:
            return False
    return True
