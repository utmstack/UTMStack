import {Directive, ElementRef, EventEmitter, HostListener, Input, OnDestroy, OnInit, Output} from '@angular/core';
import * as $ from 'jquery';
// Loads the intl-tel-input v12 jQuery plugin (UMD). It does `factory(require("jquery"), …)`
// and attaches `$.fn.intlTelInput`. Webpack dedupes that `require("jquery")` to the SAME
// jquery module as the `import * as $ from 'jquery'` above, so the plugin patches the
// jQuery instance this directive calls. (The `utils` country-data file is NOT imported
// here — it ships as a classic <script> global `intlTelInputUtils`, see src/index.html,
// and its webpack import is stubbed in webpack.custom.js for the S13 strict-scope reason.)
import 'intl-tel-input';

/**
 * Ivy-native stand-in for the View-Engine-only `ng2-tel-input` (ng2-tel-input
 * 2.3.0 has no Ivy release and Angular 17 removed ngcc, so it cannot compile
 * at 17). Replicates the exact public API + event semantics the app uses:
 * selector `[ng2TelInput]`, input `ng2TelInputOptions`, outputs
 * `countryChange`, `hasError`, `ng2TelOutput`, `intlTelInputObject`. It drives
 * the same underlying `intl-tel-input` v12 jQuery plugin (delivered as a
 * classic script; `intlTelInputUtils` is already a global, see src/index.html),
 * so the AD notification template is unchanged.
 *
 * NOTE (parity): like ng2-tel-input, `hasError` emits the *validity* boolean
 * (true = valid) on blur — the consumer (`hasErrorPhone`) stores it in
 * `invalidPhone`, which gates the "Add Phone" button. Do not "fix" the naming.
 */
@Directive({
  selector: '[ng2TelInput]',
})
export class UtmTelInputDirective implements OnInit, OnDestroy {
  @Input() ng2TelInputOptions: any;
  @Output() hasError = new EventEmitter<boolean>();
  @Output() ng2TelOutput = new EventEmitter<string>();
  @Output() countryChange = new EventEmitter<any>();
  @Output() intlTelInputObject = new EventEmitter<any>();

  private iti: any;

  constructor(private el: ElementRef) {
  }

  ngOnInit() {
    const $el = $(this.el.nativeElement);
    this.iti = $el.intlTelInput(this.ng2TelInputOptions || {});
    $el.on('countrychange', (_e, countryData) => this.countryChange.emit(countryData));
    this.intlTelInputObject.emit($el);
  }

  @HostListener('blur', [])
  onBlur() {
    const isInputValid = this.isInputValid();
    if (isInputValid) {
      const telOutput = this.iti.intlTelInput('getNumber');
      this.hasError.emit(isInputValid);
      this.ng2TelOutput.emit(telOutput);
    } else {
      this.hasError.emit(isInputValid);
    }
  }

  isInputValid(): boolean {
    return this.iti ? !!this.iti.intlTelInput('isValidNumber') : false;
  }

  setCountry(country: string) {
    if (this.iti) {
      this.iti.intlTelInput('setCountry', country);
    }
  }

  ngOnDestroy() {
    if (this.iti) {
      this.iti.intlTelInput('destroy');
    }
  }
}
