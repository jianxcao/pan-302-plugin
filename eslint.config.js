import pluginVue from 'eslint-plugin-vue'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import oxlint from 'eslint-plugin-oxlint'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

export default defineConfigWithVueTs(
  {
    name: 'plugin/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
  },

  {
    name: 'plugin/files-to-ignore',
    ignores: [
      '**/dist/**',
      '**/node_modules/**',
      '**/*.config.{js,ts,mjs,cjs}',
    ],
  },

  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  ...oxlint.configs['flat/recommended'],
  skipFormatting,

  {
    rules: {
      'no-console': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-empty-object-type': [
        'error',
        { allowInterfaces: 'always', allowObjectTypes: 'always' },
      ],
      '@typescript-eslint/no-unused-vars': [
        'warn',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/ban-ts-comment': 'warn',
      'vue/multi-word-component-names': 'off',
      'vue/valid-v-slot': [
        'error',
        {
          allowModifiers: true,
        },
      ],
      semi: ['error', 'never'],
      curly: ['error', 'all'],
      'vue/no-mutating-props': 'off',
    },
  },
)
