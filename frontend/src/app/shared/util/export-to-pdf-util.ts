import html2canvas from 'html2canvas';
import jsPDF from 'jspdf';
import * as moment from 'moment';

/**
 * Export to pdg html segment
 * @param htmlId Html ID container in DOM;
 * @param name Dashboard name
 */
export const captureScreen = (htmlId: string, name?: string): Promise<boolean> => {
  return new Promise<boolean>((resolve) => {
    buildPdf(htmlId, name).then(pdf => {
      const pdfName = 'UTM Vault Technology Dashboard-' + moment(new Date()).format('YYYY-MM-DD') + '.pdf';
      pdf.save(pdfName);
      resolve(true);
    });
  });
};

export function buildPdf(htmlId: string, name?: string): Promise<any> {
  return new Promise<any>((resolve) => {
    const data = document.getElementById(htmlId);
    html2canvas(data).then(canvas => {
      const imgWidth = 207;
      const imgHeight = canvas.height * imgWidth / canvas.width;
      const contentDataURL = canvas.toDataURL('image/png');
      const pdf = new jsPDF('p', 'mm', 'a4'); // A4 size page of PDF
      const logo = new Image();
      // logo.src = '/assets/img/logo_mini.png';
      logo.src = '/assets/img/report-front.png';
      pdf.setFontSize(8);
      const width = pdf.internal.pageSize.getWidth();
      const height = pdf.internal.pageSize.getHeight();
      pdf.addImage(logo, 'PNG', 0, 0, width, height);
      pdf.addPage();
      // pdf.text(10, 5, (name ? name : 'UTMStack'));
      // pdf.text(imgWidth - 25, 5, moment(new Date()).format('YYYY-MM-DD HH:MM:ss'));
      pdf.setFontSize(8);
      pdf.addImage(contentDataURL, 'PNG', 2, 10, imgWidth, imgHeight);
      resolve(pdf);
    });
  });
}

export const pdfPreview = (htmlId: string, name?: string): Promise<Blob> => {
  return new Promise<Blob>((resolve) => {
    buildPdf(htmlId, name).then(pdf => {
      resolve(pdf.output('blob'));
    });
  });
};

