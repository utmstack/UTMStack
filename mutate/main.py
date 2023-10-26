import json
import logging
import os
import time
from collections import Counter

from jinja2 import Environment, FileSystemLoader

from cloud_integrations.integration_creator import IntegrationCreator
from cloud_integrations.integration_enum import IntegrationEnum
from pipeline_generator import (
    generate_logstash_pipeline,
    create_input, create_filter
)
from util.misc import get_pipelines, get_active_pipelines, compare_dicts_in_unordered_lists

PIPELINES_PATH = "/usr/share/logstash/pipelines/"
CONF_FILE_PATH = "/usr/share/logstash/config/pipelines.yml"
TEMPLATE_DIR = "templates"
LOG_FORMAT = '%(asctime)s %(clientip)-15s %(user)-8s %(message)s'
SLEEP_TIME_CONFIG_GEN = 30
SLEEP_TIME_ERROR = 15

ENVIRONMENT = Environment(loader=FileSystemLoader(os.path.join(os.path.dirname(__file__), TEMPLATE_DIR)))

# Setting up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s,%(msecs)03d %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s',
    datefmt='%Y-%m-%d:%H:%M:%S')
logger = logging.getLogger(__name__)

def handle_new_pipeline(pipeline_conf):
    try:
        generate_logstash_pipeline(pipeline_root=PIPELINES_PATH,
                                   environment=ENVIRONMENT,
                                   pipeline=pipeline_conf)
    except Exception as e:
        logger.error(str(e))

def handle_pipeline_inputs(pipeline_id, pipeline_conf, last_configurations):
    try:
        current_inputs = pipeline_conf.get('inputs')
        last_inputs = last_configurations[pipeline_id].get('inputs')
        if not compare_dicts_in_unordered_lists(current_inputs, last_inputs):
            create_input(
                pipeline_directory=os.path.join(PIPELINES_PATH, pipeline_id),
                pipeline_id=pipeline_id,
                inputs=current_inputs,
                environment=ENVIRONMENT
            )
    except Exception as e:
        logger.error(str(e))


def handle_pipeline_filters(pipeline_id, pipeline_conf, last_configurations):
    try:
        current_filters = pipeline_conf.get('filters')
        last_filters = last_configurations[pipeline_id].get('filters')
        if Counter(current_filters) != Counter(last_filters):
            create_filter(os.path.join(PIPELINES_PATH, pipeline_id), current_filters)
    except Exception as e:
        logger.error(str(e))


def check_and_update_configurations(last_configurations, current_configurations):
    """Check and update configurations if there are changes."""
    for pipeline_id, pipeline_conf in current_configurations.items():
        if pipeline_id not in last_configurations.keys():
            handle_new_pipeline(pipeline_conf)
        else:
            handle_pipeline_inputs(pipeline_id, pipeline_conf, last_configurations)
            handle_pipeline_filters(pipeline_id, pipeline_conf, last_configurations)


def check_and_update_cloud_integrations(last_cloud_integrations, current_cloud_integrations, current_active_pipelines):
    """Checks and updates cloud integrations if there are changes."""
    for pipeline_id in current_cloud_integrations.keys():
        try:
            if pipeline_id in current_active_pipelines and current_cloud_integrations[pipeline_id] != None and ((last_cloud_integrations[pipeline_id] == None) or (current_cloud_integrations[pipeline_id] != last_cloud_integrations[pipeline_id])):
                logger.info("Creating {} input...".format(pipeline_id))
                create_input(
                    pipeline_directory=os.path.join(PIPELINES_PATH, pipeline_id),
                    pipeline_id=pipeline_id,
                    inputs=current_cloud_integrations[pipeline_id],
                    environment=ENVIRONMENT
                )
        except Exception as e:
            logger.error(str(e))


def check_and_update_active_pipelines(last_active_pipelines, current_active_pipelines):
    """Checks and updates active pipelines if there are changes."""

    try:
        if Counter(last_active_pipelines) == Counter(current_active_pipelines):
            return
        generate_logstash_config(current_active_pipelines)

    except Exception as e:
        logger.error(str(e))


def generate_logstash_config(pipelines):
    """Generates the configuration file for logstash using the provided pipeline information."""
    try:
        template = ENVIRONMENT.get_template('pipelines.yml.j2')

        active_pipelines = {
            pipeline_id: os.path.join(PIPELINES_PATH, pipeline_id)
            for pipeline_id in pipelines
            if pipeline_id is not None
        }

        content = template.render(active_pipelines=active_pipelines)

        with open(CONF_FILE_PATH, "w", encoding='utf-8') as file:
            file.write(content)
    except Exception as e:
        logger.error(str(e))


def main():
    """Main loop for periodically updating Logstash configuration."""
    
    logger.info("Starting Mutate...")
    
    last_configurations = {}
    last_cloud_integrations = {
        'cloud_google': None,
        'cloud_azure': None
    }
    last_active_pipelines = []

    while True:
        try:
            logger.info("Configuring Pipelines...")

            current_configurations = get_pipelines()
            logger.info("Pipelines configurations obtained correctly")

            current_cloud_integrations = {
                'cloud_google': IntegrationCreator().create_integration(
                    IntegrationEnum.GOOGLE
                ).get_integration_config(),
                'cloud_azure': IntegrationCreator().create_integration(
                    IntegrationEnum.AZURE
                ).get_integration_config()
            }
            logger.info("Cloud Integrations configurations obtained correctly")

            current_active_pipelines = get_active_pipelines()
            logger.info("Active Pipelines obtained correctly")

            check_and_update_configurations(last_configurations, current_configurations)
            logger.info("Pipelines configurations checked and updated correctly")

            check_and_update_cloud_integrations(last_cloud_integrations, current_cloud_integrations, current_active_pipelines)
            logger.info("Cloud Integrations configurations checked and updated correctly")

            check_and_update_active_pipelines(last_active_pipelines, current_active_pipelines)
            logger.info("Active Pipelines checked and updated correctly")

            last_configurations = current_configurations
            last_cloud_integrations = current_cloud_integrations
            last_active_pipelines = current_active_pipelines

            time.sleep(SLEEP_TIME_CONFIG_GEN)
        except Exception as e:
            logger.exception(e)
            time.sleep(SLEEP_TIME_ERROR)


if __name__ == "__main__":
    main()


