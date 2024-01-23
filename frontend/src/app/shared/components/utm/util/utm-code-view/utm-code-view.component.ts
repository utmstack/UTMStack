import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
    selector: 'app-utm-code-view',
    templateUrl: './utm-code-view.component.html',
    styleUrls: ['./utm-code-view.component.scss']
})
export class UtmCodeViewComponent implements OnInit, OnChanges {
    @Input() code: string;
    @Input() allowCopy = true;
    private copy: any;
    str: string;

    constructor(private utmToastService: UtmToastService) {
    }

    ngOnInit() {
        this.str = this.code;

    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes.code) {
            this.str = this.code;
            this.code = this.maskSecrets(this.code);
        }
    }

    preventCopy(event: ClipboardEvent): void {
        event.preventDefault();
    }

    copyCode() {
        const selBox = document.createElement('textarea');
        const copyText = this.removeSecretsTags(this.str.split('<br>').join(''));
        selBox.style.position = 'fixed';
        selBox.style.left = '0';
        selBox.style.top = '0';
        selBox.style.opacity = '0';
        selBox.value = copyText;
        document.body.appendChild(selBox);
        selBox.focus();
        selBox.select();
        document.execCommand('copy');

        document.body.removeChild(selBox);
        this.utmToastService.showInfo('Copied!', 'Code is in clipboard');
    }


    maskSecrets(str) {
        const regex = /<secret>(.*?)<\/secret>/gs;
        return str.replace(regex, (_, match) => '*'.repeat(match.length));
    }

    removeSecretsTags(str) {
        const regex = /<\/?secret>/g;
        return str.replace(regex, '');
    }
}
