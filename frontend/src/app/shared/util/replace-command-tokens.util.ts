export function replaceCommandTokens(command: string, wordsToReplace: { [key: string]: string }) {
  return Object.keys(wordsToReplace)
    .reduce((f, s) => f.replace(new RegExp(s, 'ig'), wordsToReplace[s]), command);
}
