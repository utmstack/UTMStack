/*export function replaceCommandTokens(command: string, wordsToReplace: { [key: string]: string }) {
  return Object.keys(wordsToReplace)
    .reduce((f, s) => f.replace(new RegExp(s, 'ig'), wordsToReplace[s]), command);
}*/

export function replaceCommandTokens(command: string, wordsToReplace: { [key: string]: string }) {
  let cmd = command;

  Object.entries(wordsToReplace).forEach(([key, value]) => {
    if (!value) {
      const regex = new RegExp(`\\s*${key}\\b`, 'g');
      cmd = cmd.replace(regex, '');
    } else {
      const regex = new RegExp(`${key}\\b`, 'g');
      cmd = cmd.replace(regex, value);
    }
  });

  cmd = cmd.replace(/\s+/g, ' ').trim();

  return cmd;
}
