import pluginVue from "eslint-plugin-vue";
import parserVue from "vue-eslint-parser";

export default [
  {
    files: ["**/*.vue"],
    languageOptions: {
      parser: parserVue,
      ecmaVersion: "latest",
      sourceType: "module",
    },
    plugins: {
      vue: pluginVue,
    },
    rules: {
      ...pluginVue.configs["vue3-essential"].rules,
      "vue/multi-word-component-names": "off",
      "vue/no-v-html": "off",
      "no-unused-vars": [
        "warn",
        {
          vars: "all",
          args: "after-used",
          ignoreRestSiblings: true,
          varsIgnorePattern: "^_",
          argsIgnorePattern: "^_",
        },
      ],
      "no-console": "warn",
    },
  },
  {
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: {
        browser: true,
        node: true,
      },
    },
    rules: {
      "no-unused-vars": [
        "warn",
        {
          vars: "all",
          args: "after-used",
          ignoreRestSiblings: true,
          varsIgnorePattern: "^_",
          argsIgnorePattern: "^_",
        },
      ],
      "no-console": "warn",
    },
  },
];
