import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from "rxjs";
import { delay } from 'rxjs/operators';
import {Rule} from '../models/rule.model';

@Injectable()
export class RuleService {
    rules: Rule[] = [
        {
            id: 200001,
            data_types: ['linux'],
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
                    { get: 'local.host', as: 'hostname', of_type: 'string' },
                    { get: 'dataType', as: 'dataType', of_type: 'string' }
                ],
                expression: 'dataType == "linux"'
            }
        }
    ];

    constructor(private http: HttpClient) {
    }

    loadAll(): Observable<Rule[]>{
       return  of(this.rules)
           .pipe(
               delay(3000)
           );
    }
}
