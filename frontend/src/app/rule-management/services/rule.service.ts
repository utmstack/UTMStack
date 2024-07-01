import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, delay} from 'rxjs/operators';
import {Rule} from '../models/rule.model';

@Injectable()
export class RuleService {
    rules: Rule[] = [
        {
            id: 200001,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        },
        {
            id: 200002,
            data_types: [{
                id: 2, name: 'windows'
            }],
            name: 'Testing Windows rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        },
        {
            id: 200003,
            data_types: [{
                id: 3, name: 'winevenlog'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        },
        {
            id: 200004,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200005,
            data_types: [{
                id: 4, name: 'netflow'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200006,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200007,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200008,
            data_types: [{
                id: 5, name: 'syslog'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200009,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200010,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 200011,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 2000012,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 2000013,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 2000014,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 2000015,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com', 'http://utmstack.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }, {
            id: 2000016,
            data_types: [{
                id: 1, name: 'linux'
            }],
            name: 'Testing Linux rule',
            impact: {
                confidentiality: 3,
                integrity: 1,
                availability: 3
            },
            category: 'Testing',
            technique: 'Testing technique',
            description: 'This is a test rule for Linux logs',
            references: ['http://threatwinds.com'],
            definition: {
                variables: [
                    {get: 'local.host', as: 'hostname', of_type: 'string'},
                    {get: 'dataType', as: 'dataType', of_type: 'string'}
                ],
                expression: 'dataType == "linux"'
            }
        }
    ];

    constructor(private http: HttpClient) {
    }

    getRules(): Observable<HttpResponse<Rule[]>> {
        return of(new HttpResponse({
            headers: new HttpHeaders({'X-Total-Count': this.rules.length.toString()}),
            body: this.rules
        })).pipe(
            delay(2000),
            catchError((error: any): Observable<HttpResponse<Rule[]>> => {
                console.error('Error loading rules', error);
                return of(new HttpResponse({
                    headers: new HttpHeaders({'X-Total-Count': '0'}),
                    body: []
                }));
            })
        );
    }

    getRuleById(id: number): Observable<HttpResponse<Rule>> {
       return of(new HttpResponse({
           body: this.rules.find(rule => rule.id === id)
       })).pipe(
               delay(2000),
               catchError((error: any): Observable<HttpResponse<Rule>> => {
                   console.error('Error loading rules', error);
                   return of(new HttpResponse({
                       body: undefined
                   }));
               })
           );
    }

    getFieldValue(params: any) {
        return of(new HttpResponse({
            body: this.getUniqueValues(params.prop)
        })).pipe(
            delay(2000),
            catchError((error: any): Observable<HttpResponse<Array<[string, number]>>> => {
                console.error('Error loading rules', error);
                return of(new HttpResponse({
                    body: undefined
                }));
            })
        );
    }

    private getUniqueValues(attributeName: string): Array<[string, number]> {
        const uniqueValues: Set<string> = new Set();
        const uniqueValuesArray: Array<[string, number]> = [];

        for (const rule of this.rules) {
            const attributeValue = this.getValueFromPath(rule, attributeName);

            if (Array.isArray(attributeValue)) {
                attributeValue.forEach( a => {
                    if (typeof a === 'object') {
                        uniqueValues.add(a.name);
                    } else {
                        uniqueValues.add(a);
                    }

                });
            }

            if (attributeValue !== undefined &&  !Array.isArray(attributeValue)) {
                uniqueValues.add(attributeValue.toString());
            }
        }

        Array.from(uniqueValues).forEach((value, index) =>
            uniqueValuesArray.push([value, index]));

        return uniqueValuesArray;
    }
    private getValueFromPath(obj: any, path: string): any {
        const parts = path.split('.');
        let value = obj;
        for (const part of parts) {
            if (value[part] === undefined) {
                return undefined;
            }
            value = value[part];
        }
        return value;
    }
}
