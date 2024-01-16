# pylint: disable=R1732
"""UTMStack utilities."""

import json
import logging.handlers
from typing import Any, Dict

# pylama:ignore=W0611
from util.postgres import Postgres

logging.basicConfig(level=logging.INFO, format='%(asctime)s,%(msecs)03d %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s',
    datefmt='%Y-%m-%d:%H:%M:%S')
logger = logging.getLogger(__name__)

def get_module_group(module: str):
    """Get groups of configuration module"""
    query = """select distinct group_name from
    utm_server_configurations WHERE module_name=%s;"""
    queryresult = Postgres().fetchall(query, (module,))
    groups = [group['group_name'] for group in queryresult]
    return groups


def get_config(module: str, group: str) -> Dict[str, Any]:
    """Get configuration for the given module."""
    query = """SELECT conf_short, conf_value
        FROM utm_server_configurations
        WHERE module_name=%s AND group_name=%s;"""
    configs = Postgres().fetchall(
        query, (module, group))
    cfg = {}
    for row in configs:
        key = row['conf_short']
        value_str = row['conf_value']
        try:
            value = json.loads(value_str)
        except Exception:
            value = value_str
        cfg[key] = value
    return cfg


def get_pipelines():
    try:
        query_result = Postgres().fetchall("""
        select
            ulg.id,
            ulp.pipeline_id,
            json_agg(json_build_object('input_plugin', uli.input_plugin, 'conf', confs))
                  as inputs,
            json_agg(distinct ulg.logstash_filter) as filters
        from
            utm_logstash_pipeline as ulp left join utm_logstash_input as uli on ulp.id = uli.pipeline_id
        left join ( select ulic.id, ulic.input_id, json_build_object(ulic.conf_key, ulic.conf_value) as confs
            from utm_logstash_input_configuration as ulic group by ulic.id, ulic.input_id, ulic.conf_key, ulic.conf_value
            order by ulic.id ) as confs on uli.id = confs.input_id
        left join  ( select ulg.*, ulf.logstash_filter from utm_group_logstash_pipeline_filters ulg left join utm_logstash_filter ulf
            on ulg.filter_id = ulf.id order by id asc) as ulg 
            on ulg.pipeline_id = ulp.id
        group by
            ulp.pipeline_id,
            ulg.id;
        """)

        pipelines_dict = {}
        for row in query_result:
            pipeline_id = row['pipeline_id']
            if pipeline_id not in pipelines_dict:
                pipelines_dict[pipeline_id]= {
                    'pipeline_id': pipeline_id,
                    'inputs': row['inputs'],
                    'filters': row['filters']
                    }
            else:
                pipelines_dict[pipeline_id]['filters'].extend(row['filters'])
        
        return pipelines_dict

    except Exception as e:
        raise


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
        raise


def compare_dicts_in_unordered_lists(list1, list2):
    if len(list1) != len(list2):
        return False
    for dictionary in list1:
        if dictionary not in list2:
            return False
    return True
