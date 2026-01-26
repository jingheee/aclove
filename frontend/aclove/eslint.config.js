{
  "root": true,
  "env": {
    "browser": true,
    "es2022": true,
    "node": true
  },
  "extends": [
    "eslint:recommended",
    "plugin:vue/vue3-essential"
  ],
  "parserOptions": {
    "ecmaVersion": "latest",
    "sourceType": "module"
  },
  "plugins": {
    "vue": "plugin:vue/essential"
  },
  "rules": {
    "vue/multi-word-component-names": "off",
    "vue/no-v-html": "off",
    "no-unused-vars": "warn",
    "no-console": "warn"
  }
}
