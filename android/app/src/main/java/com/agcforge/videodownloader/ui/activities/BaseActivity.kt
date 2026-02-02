package com.agcforge.videodownloader.ui.activities

import android.content.Context
import android.content.res.Configuration
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.applyTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import java.util.Locale

abstract class BaseActivity : AppCompatActivity() {

    private lateinit var preferenceManager: PreferenceManager
	private var appliedLanguageCode: String? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        preferenceManager = PreferenceManager(this)
        observeTheme()
        observeLanguage()
        super.onCreate(savedInstanceState)
    }

    private fun observeTheme() {
        lifecycleScope.launch {
            preferenceManager.theme.first().let { theme ->
                applyTheme(theme)
            }
        }
    }
    private fun observeLanguage() {
        lifecycleScope.launch {
			preferenceManager.language
				.distinctUntilChanged()
				.collect { languageCode ->
					if (appliedLanguageCode == null) {
						appliedLanguageCode = languageCode
						return@collect
					}
					if (appliedLanguageCode != languageCode) {
						appliedLanguageCode = languageCode
						recreate()
					}
				}
        }
    }

    override fun attachBaseContext(newBase: Context) {
        preferenceManager = PreferenceManager(newBase)
        val languageCode = runBlocking { preferenceManager.language.first() }
		appliedLanguageCode = languageCode
        val context = updateBaseContextLocale(newBase, languageCode)
        super.attachBaseContext(context)
    }

    private fun updateBaseContextLocale(context: Context, languageCode: String?): Context {
        val locale = if (!languageCode.isNullOrEmpty()) {
            Locale(languageCode)
        } else {
            // If no language is saved, use the system default
            Locale.getDefault()
        }
        Locale.setDefault(locale)
        val config = Configuration(context.resources.configuration)
        config.setLocale(locale)
        return context.createConfigurationContext(config)
    }

    fun restartActivity() {
		recreate()
    }
}
