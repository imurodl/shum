import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import globals from 'globals'

export default [
  {
    ignores: ['dist/**', 'node_modules/**'],
  },
  js.configs.recommended,
  // 'essential' catches real bugs; Prettier owns formatting, so we skip the
  // stylistic vue rulesets that would fight it.
  ...pluginVue.configs['flat/essential'],
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
      },
    },
    rules: {
      // v-html is used intentionally for author-controlled i18n/highlighted code.
      'vue/no-v-html': 'off',
      'vue/multi-word-component-names': 'off',
    },
  },
]
