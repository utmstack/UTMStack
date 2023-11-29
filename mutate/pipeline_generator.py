import logging
import os

logging.basicConfig(level=logging.INFO,
                    format="%(asctime)s - %(levelname)s - %(message)s",
                    datefmt="%d-%b-%y %H:%M:%S")


def generate_logstash_pipeline(pipeline_root, environment, pipeline: dict) -> None:
    """
    Generates the complete pipeline by invoking helper methods to create folders,
    inputs, filters, and outputs.

    :param active_pipelines:
    :param pipeline_root: The root directory for the pipelines.
    :param environment: Enviroment of the templates.
    :param pipeline: Dictionary containing pipeline data.
    """
    pipeline_directory = create_directory(pipeline_root, pipeline['pipeline_id'])
    create_input(pipeline_directory, pipeline['pipeline_id'], pipeline['inputs'], environment)
    create_filter(pipeline_directory, pipeline['filters'])
    create_output(pipeline_directory, environment)


def create_directory(root_dir, directory_name):
    """
    Creates a new directory under the root directory.

    :param root_dir: The root directory under which the new directory is to be created.
    :param directory_name: The name of the new directory.
    :return: The path of the created directory.
    """
    new_directory_path = os.path.join(root_dir, directory_name)
    if not os.path.exists(new_directory_path):
        try:
            os.makedirs(new_directory_path)
        except OSError as error:
            logging.error(f"Unable to create directory '{new_directory_path}'. Error: {error}")
    return new_directory_path


def create_input(pipeline_directory, pipeline_id, inputs, environment):
    """
    Generates the input configuration file.

    :param pipeline_directory: The directory where the input file is to be generated.
    :param pipeline_id: The id of the pipeline.
    :param inputs: The list of inputs.
    :param environment: The Jinja environment for the use of the template.
    """

    path = os.path.join(pipeline_directory, "000-input.conf")
    if pipeline_id in ['cloud_azure', 'cloud_google'] and isinstance(inputs, str):
        with open(path, "w", errors="ignore", encoding='utf-8') as file:
            file.write("input{" + inputs + "}")
    else:
        inputs_content = ""
        for input_item in inputs:
            try:
                input_plugin = input_item.get('input_plugin', None)
                if input_plugin is None:
                    continue

                template_name = f"{input_plugin}_template.j2"

                if not os.path.isfile(os.path.join(os.path.dirname(__file__), "templates", template_name)):
                    logging.error(f"No template exists for the input plugin: {input_plugin}")
                    continue

                template = environment.get_template(template_name)

                configurations = input_item.get('conf', {})

                content = template.render(**configurations)
                inputs_content += content

            except Exception as e:
                logging.error(f"Error during input file generation: {e}")
                continue

        if inputs_content:
            with open(path, "w", errors="ignore", encoding='utf-8') as file:
                file.write("input{" + inputs_content + "}")


def create_filter(pipeline_directory, filters):
    """
    Generates the filter configuration file.

    :param pipeline_directory: The directory where the filter file is to be generated.
    :param filters: The list of filters.
    """
    try:
        if not filters or filters[0] is None:
            return

        filters_content = "\n\n".join(filters)

        path = os.path.join(pipeline_directory, "111-filter.conf")
        with open(path, "w", errors="ignore", encoding='utf-8') as file:
            file.write(filters_content)

    except Exception as e:
        logging.error(f"Error during filter file generation: {e}")


def create_output(pipeline_directory, environment):
    """
    Generates the output configuration file.

    :param active_pipelines:
    :param pipeline_directory: The directory where the output file is to be generated.
    :param environment: The environment where the pipeline is to be generated.
    """

    template_name = 'output_template.j2'

    try:
        template = environment.get_template(template_name)
        content = template.render(correlation_url=os.environ.get('CORRELATION_URL'))

        path = os.path.join(pipeline_directory, "999-output.conf")
        with open(path, "w", errors="ignore", encoding='utf-8') as file:
            file.write(content)

    except Exception as e:
        logging.error(f"Error during output file generation: {e}")