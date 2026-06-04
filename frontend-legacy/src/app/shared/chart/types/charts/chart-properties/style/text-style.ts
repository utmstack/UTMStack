export class TextStyle {
  fontSize?: number;
  fontFamily?: string;


  constructor(fontSize?: number, fontFamily?: string) {
    this.fontSize = fontSize ? fontSize : 12;
    this.fontFamily = fontFamily ? fontFamily : '"Poppins", sans-serif;';
  }
}
