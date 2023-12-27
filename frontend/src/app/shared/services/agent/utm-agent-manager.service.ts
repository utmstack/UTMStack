import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {AgentCommandType, AgentType} from '../../types/agent/agent.type';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})

export class UtmAgentManagerService {
  public resourceUrl = SERVER_API_URL + 'api/agent-manager';

  constructor(private http: HttpClient) {
  }


  getAgents(req?: any): Observable<HttpResponse<AgentType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AgentType[]>(this.resourceUrl + '/agents', {params: options, observe: 'response'});
  }

  getAgentsWithCommands(req?: any): Observable<HttpResponse<AgentType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AgentType[]>(this.resourceUrl + '/agents-with-commands', {params: options, observe: 'response'});
  }

  getAgent(hostname: string): Observable<HttpResponse<AgentType>> {
    return this.http.get<AgentType>(this.resourceUrl + '/agent-by-hostname?hostname=' + hostname, {
      observe: 'response'
    });
  }

  canRunCommandOnAgent(hostname: string): Observable<HttpResponse<AgentType>> {
    return this.http.get<AgentType>(this.resourceUrl + '/can-run-command?hostname=' + hostname, {
      observe: 'response'
    });
  }

  getAgentCommands(req?: any): Observable<HttpResponse<AgentCommandType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AgentCommandType[]>(this.resourceUrl + '/agent-commands', {params: options, observe: 'response'});
  }


}
