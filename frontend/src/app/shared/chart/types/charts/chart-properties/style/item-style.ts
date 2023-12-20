export class ItemStyle {
  normal?: {
    borderWidth?: number,
    borderColor?: string,
    label?: {
      show?: boolean,
      position?: any,
      padding?: any[],
      fontSize?: number
      rotate?: any,
      fontWeight?: any,
      color?: string,
      formatter?: any
    },
    labelLine?: {
      show?: boolean
    },
    areaStyle?: {
      type?: 'mint',
      normal?: {
        opacity?: number
      }
    }
  };
  emphasis?: {
    label?: {
      show?: boolean,
      position?: any,
      textStyle?: {
        fontSize?: string,
        fontWeight?: string
      }
    }
  };

  constructor(normal?: {
                borderWidth?: number,
                borderColor?: string,
                label?: {
                  show?: boolean,
                  position?: any,
                  rotate?: any,
                  padding?: any[],
                  fontSize?: number,
                  color?: string,
                  fontWeight?: any,
                  formatter?: any
                },
                labelLine?: { show?: boolean },
                areaStyle?: {
                  type?: 'mint',
                  normal?: {
                    opacity?: number
                  }
                }
              },
              emphasis?: {
                label?: {
                  show?: boolean,
                  position?: 'center' | 'top' | 'bottom',
                  textStyle?: { fontSize?: string, fontWeight?: string }
                }
              }) {
    this.normal = normal ? normal : null;
    this.emphasis = emphasis ? emphasis : null;
  }
}
