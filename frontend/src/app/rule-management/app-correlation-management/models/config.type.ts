export type Routes = 'assets' | 'types' | 'patterns';

export const RouteValues: { [key in Routes]: Routes } = {
    assets: 'assets',
    types: 'types',
    patterns: 'patterns',
};

export enum Actions {
    CREATE_ASSET = 'Add asset',
    CREATE_TYPE = 'Add data type',
    CREATE_PATTERN = 'Add pattern',
    CREATE_RULE = 'Add rule'
}
