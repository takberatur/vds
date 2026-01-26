package com.agcforge.videodownloader.ui.activities

import android.content.Context
import android.content.res.Configuration
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.applyTheme
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import java.util.Locale

abstract class BaseActivity : AppCompatActivity() {

    private lateinit var preferenceManager: PreferenceManager

    override fun onCreate(savedInstanceState: Bundle?) {
        preferenceManager = PreferenceManager(this)
        observeTheme()
        super.onCreate(savedInstanceState)
    }

    private fun observeTheme() {
        lifecycleScope.launch {
            preferenceManager.theme.first().let { theme ->
                applyTheme(theme)
            }
        }
    }

    override fun attachBaseContext(newBase: Context) {
        preferenceManager = PreferenceManager(newBase)
        val languageCode = runBlocking { preferenceManager.language.first() }
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
}
