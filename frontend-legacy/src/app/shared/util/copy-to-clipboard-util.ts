/**
 * Copy text to clipboard
 * @param text Text to copy;
 */
export const copyToClipboard = (text: string): Promise<boolean> => {
  return new Promise<boolean>((resolve) => {
    const selBox = document.createElement('textarea');
    selBox.style.position = 'fixed';
    selBox.style.left = '0';
    selBox.style.top = '0';
    selBox.style.opacity = '0';
    selBox.value = text;
    document.body.appendChild(selBox);
    selBox.focus();
    selBox.select();
    document.execCommand('copy');
    document.body.removeChild(selBox);
    resolve(true);
  });
};

