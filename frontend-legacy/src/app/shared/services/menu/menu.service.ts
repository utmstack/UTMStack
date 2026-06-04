import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {IMenu} from '../../types/menu/menu.model';
import {QueryType} from '../../types/query-type';

@Injectable({
  providedIn: 'root'
})
export class MenuService {

  public resourceUrl = SERVER_API_URL + 'api/menu';

  public loadMenu: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(true);
  public loadSubMenu: BehaviorSubject<any> = new BehaviorSubject<boolean>(null);

  constructor(private http: HttpClient) {
  }

  /**
   * Create
   * @param menu : Menu to create
   */
  create(menu: IMenu): Observable<HttpResponse<IMenu>> {
    return this.http.post<IMenu>(this.resourceUrl, menu, {observe: 'response'});
  }

  /**
   * Update
   * @param menu : Menu to update
   */
  update(menu: IMenu): Observable<HttpResponse<IMenu>> {
    return this.http.put<IMenu>(this.resourceUrl, menu, {observe: 'response'});
  }

  // POST /api/menu/save-menu-structure
  saveStructure(menus: IMenu[]): Observable<HttpResponse<IMenu>> {
    return this.http.post<IMenu>(this.resourceUrl + '/save-menu-structure', menus, {observe: 'response'});
  }

  /**
   * Find menus by name
   * @param name : Name to search
   */
  find(name: string): Observable<HttpResponse<IMenu>> {
    return this.http.get<IMenu>(`${this.resourceUrl}/${name}`, {observe: 'response'});
  }

  /**
   * Find menus by filters
   */
  query(query?: QueryType): Observable<HttpResponse<IMenu[]>> {
    return this.http.get<IMenu[]>(this.resourceUrl.concat(query ? query.toString() : ''), {observe: 'response'});
  }

  /**
   * Find menus by filters
   */
  getMenuStructure(includeModulesMenus: boolean): Observable<HttpResponse<IMenu[]>> {
    return this.http.get<IMenu[]>(this.resourceUrl + '/all?includeModulesMenus=' + includeModulesMenus, {observe: 'response'});
  }

  /**
   * Delete menu
   * @param id : Menu identifier to delete
   */
  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  /**
   * Delete menu by URL
   * @param url Menu url
   */

  deleteByUrl(url: string): Observable<HttpResponse<any>> {
    return this.http.delete(this.resourceUrl + '/delete-by-url?url=' + url, {observe: 'response'});
  }

  findMenusByAuthorities(query: QueryType): Observable<HttpResponse<IMenu[]>> {
    return this.http.get<IMenu[]>(this.resourceUrl.concat('/get-menu-by-authorities', query ? query.toString() : ''),
      {observe: 'response'});
  }
}
