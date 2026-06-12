/* Operators for a rule condition row. Mirrors backend FilterOperator
   (pkg/common_models/filter_type.go). `needs` controls the value editor:
   - none: no value (EXIST / DOES_NOT_EXIST)
   - one:  single text input
   - list: comma-separated values for the *_ONE_OF / IS_BETWEEN ranges
*/
export type OperatorNeeds = 'none' | 'one' | 'list'

export interface TaggingOperator {
  id: string
  label: string
  needs: OperatorNeeds
}

export const RULE_OPERATORS: TaggingOperator[] = [
  { id: 'IS', label: 'is', needs: 'one' },
  { id: 'IS_NOT', label: 'is not', needs: 'one' },
  { id: 'CONTAIN', label: 'contains', needs: 'one' },
  { id: 'DOES_NOT_CONTAIN', label: 'does not contain', needs: 'one' },
  { id: 'START_WITH', label: 'starts with', needs: 'one' },
  { id: 'NOT_START_WITH', label: 'does not start with', needs: 'one' },
  { id: 'ENDS_WITH', label: 'ends with', needs: 'one' },
  { id: 'NOT_ENDS_WITH', label: 'does not end with', needs: 'one' },
  { id: 'IS_ONE_OF', label: 'is one of', needs: 'list' },
  { id: 'IS_NOT_ONE_OF', label: 'is not one of', needs: 'list' },
  { id: 'CONTAIN_ONE_OF', label: 'contains one of', needs: 'list' },
  { id: 'DOES_NOT_CONTAIN_ONE_OF', label: 'does not contain one of', needs: 'list' },
  { id: 'IS_BETWEEN', label: 'is between', needs: 'list' },
  { id: 'IS_NOT_BETWEEN', label: 'is not between', needs: 'list' },
  { id: 'EXIST', label: 'exists', needs: 'none' },
  { id: 'DOES_NOT_EXIST', label: 'does not exist', needs: 'none' },
]

export function operatorById(id: string): TaggingOperator | undefined {
  return RULE_OPERATORS.find((o) => o.id === id)
}

// Fields a tagging rule may match on. The legacy app intentionally excluded
// triage/audit fields (status, notes, history, …); the catalog below mirrors
// that — anything the user can usefully match in raw alert data.
export const RULE_FIELDS: { label: string; field: string }[] = [
  { label: 'Alert name', field: 'name' },
  { label: 'Category', field: 'category' },
  { label: 'Technique', field: 'technique' },
  { label: 'Datasource', field: 'dataSource' },
  { label: 'Datasource group', field: 'assetGroupName' },
  { label: 'Data type', field: 'dataType' },
  { label: 'Protocol', field: 'protocol' },
  { label: 'Adversary IP', field: 'adversary.ip' },
  { label: 'Adversary host', field: 'adversary.host' },
  { label: 'Adversary user', field: 'adversary.user' },
  { label: 'Adversary domain', field: 'adversary.domain' },
  { label: 'Adversary URL', field: 'adversary.url' },
  { label: 'Adversary country', field: 'adversary.geolocation.country' },
  { label: 'Adversary ASN', field: 'adversary.geolocation.asn' },
  { label: 'Adversary ASO', field: 'adversary.geolocation.aso' },
  { label: 'Target IP', field: 'target.ip' },
  { label: 'Target host', field: 'target.host' },
  { label: 'Target user', field: 'target.user' },
  { label: 'Target domain', field: 'target.domain' },
  { label: 'Target URL', field: 'target.url' },
  { label: 'Target country', field: 'target.geolocation.country' },
  { label: 'Target ASN', field: 'target.geolocation.asn' },
  { label: 'Target ASO', field: 'target.geolocation.aso' },
]

export const TAGGING_RULES_PAGE_SIZE = 20

export const SELECT_CLS =
  'h-9 cursor-pointer rounded-md border border-input bg-background px-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'
