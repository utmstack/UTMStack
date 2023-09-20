## Opensearch Docker Images

Main repository where Dockerfiles for Opensearch are hosted. These docker files are used to build images for Opensearch Offical Images

## Usage

To use OpenSearch Official Images from Docker hub run

> docker pull opensearch:latest

The docker images are also available on Amazon ECR. To use ECR images run

> docker pull public.ecr.aws/opensearchproject/opensearch:latest

## Building

To build the docker images, you can use the following command.

> cd {2.x} && docker build -t opensearch:{major_version} -f Dockerfile .

## Getting Help

If you find a bug, or have a feature request, please don't hesitate to open an issue in this repository.

For more information, see [project website](https://opensearch.org/) and [documentation](https://docs-beta.opensearch.org/). If you need help and are unsure where to open an issue, try [forums](https://discuss.opendistrocommunity.dev/).

## Code of Conduct

This project has adopted the [Amazon Open Source Code of Conduct](CODE_OF_CONDUCT.md). For more information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq), or contact [opensource-codeofconduct@amazon.com](mailto:opensource-codeofconduct@amazon.com) with any additional questions or comments.

## Security

If you discover a potential security issue in this project we ask that you notify AWS/Amazon Security via our [vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/). Please do **not** create a public GitHub issue.

## License

This project is licensed under the [Apache v2.0 License](LICENSE.txt).

## Copyright

Copyright OpenSearch Contributors. See [NOTICE](NOTICE) for details.


