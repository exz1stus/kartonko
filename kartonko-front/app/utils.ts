function translitCyrillic(input: string) {
  const map: { [key: string]: string } = {
    'а':'a','б':'b','в':'v','г':'g','д':'d','е':'e','ё':'e','ж':'zh','з':'z','и':'i','й':'i',
    'к':'k','л':'l','м':'m','н':'n','о':'o','п':'p','р':'r','с':'s','т':'t','у':'u','ф':'f',
    'х':'kh','ц':'ts','ч':'ch','ш':'sh','щ':'shch','ъ':'','ы':'y','ь':'','э':'e','ю':'yu','я':'ya',
    'ґ':'g','є':'ye','і':'i','ї':'i','ў':'w',
  };

  return input
    .toLowerCase()
    .split('')
    .map((char: string) => map[char] || char)
    .join('')
    .replace(/\s+/g, '_')
}

export default translitCyrillic;