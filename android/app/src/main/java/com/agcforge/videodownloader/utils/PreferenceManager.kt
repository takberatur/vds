package com.agcforge.videodownloader.utils

import android.content.Context
import android.widget.ImageView
import android.widget.Toast
import androidx.appcompat.app.AppCompatDelegate
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.agcforge.videodownloader.data.model.Application
import com.bumptech.glide.Glide
import com.bumptech.glide.load.resource.drawable.DrawableTransitionOptions
import com.google.gson.Gson
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import java.text.SimpleDateFormat
import java.util.*

// DataStore instance remains a top-level extension for easy access throughout the app
val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "app_preferences")

/**
 * Manages all application preferences using DataStore.
 * This class encapsulates the logic for saving and retrieving user data and settings.
 */
class PreferenceManager(private val context: Context) {

    companion object {
        // All preference keys are now organized within the PreferenceManager companion object
        private val TOKEN_KEY = stringPreferencesKey("auth_token")
        private val USER_ID_KEY = stringPreferencesKey("user_id")
        private val USER_EMAIL_KEY = stringPreferencesKey("user_email")
        private val USER_NAME_KEY = stringPreferencesKey("user_name")
        private val USER_AVATAR_KEY = stringPreferencesKey("user_avatar")
        private val THEME_KEY = stringPreferencesKey("theme_mode")
        private val LANGUAGE_KEY = stringPreferencesKey("language_code")
		private val STORAGE_LOCATION_KEY = stringPreferencesKey("storage_location")
		private val APPLICATION_KEY = stringPreferencesKey("application_config")
    }

    // --- Flows to observe preference changes ---

    val authToken: Flow<String?> = context.dataStore.data.map { it[TOKEN_KEY] }
    val userId: Flow<String?> = context.dataStore.data.map { it[USER_ID_KEY] }
    val userEmail: Flow<String?> = context.dataStore.data.map { it[USER_EMAIL_KEY] }
    val userName: Flow<String?> = context.dataStore.data.map { it[USER_NAME_KEY] }
    val userAvatar: Flow<String?> = context.dataStore.data.map { it[USER_AVATAR_KEY] }
    val theme: Flow<String?> = context.dataStore.data.map { it[THEME_KEY] }
    val language: Flow<String?> = context.dataStore.data.map { it[LANGUAGE_KEY] }
	val storageLocation: Flow<String?> = context.dataStore.data.map { it[STORAGE_LOCATION_KEY] }
	val applicationConfig: Flow<String?> = context.dataStore.data.map { it[APPLICATION_KEY] }

    // --- Suspend functions to modify preferences ---

    suspend fun saveAuthToken(token: String) {
        context.dataStore.edit { it[TOKEN_KEY] = token }
    }

    suspend fun saveUserInfo(userId: String, email: String, name: String) {
        context.dataStore.edit { preferences ->
            preferences[USER_ID_KEY] = userId
            preferences[USER_EMAIL_KEY] = email
            preferences[USER_NAME_KEY] = name
        }
    }

    suspend fun saveUserProfile(name: String, avatarUrl: String?) {
        context.dataStore.edit { preferences ->
            preferences[USER_NAME_KEY] = name
            avatarUrl?.let { preferences[USER_AVATAR_KEY] = it }
        }
    }

    suspend fun saveTheme(themeMode: String) {
        context.dataStore.edit { it[THEME_KEY] = themeMode }
    }

    suspend fun saveLanguage(languageCode: String) {
        context.dataStore.edit { it[LANGUAGE_KEY] = languageCode }
    }

	suspend fun saveStorageLocation(location: String) {
		context.dataStore.edit { it[STORAGE_LOCATION_KEY] = location }
	}

	suspend fun saveApplication(app: Application) {
		val json = Gson().toJson(app)
		context.dataStore.edit { it[APPLICATION_KEY] = json }
	}

    suspend fun clearUserData() {
        context.dataStore.edit { it.clear() }
    }
}

// --- General-purpose Utility Functions ---

fun Context.showToast(message: String, duration: Int = Toast.LENGTH_SHORT) {
    Toast.makeText(this, message, duration).show()
}

fun ImageView.loadImage(url: String?, placeholder: Int = android.R.drawable.ic_menu_gallery) {
    Glide.with(this.context)
        .load(url)
        .placeholder(placeholder)
        .error(placeholder)
        .transition(DrawableTransitionOptions.withCrossFade())
        .into(this)
}

fun String.formatDate(inputFormat: String = "yyyy-MM-dd'T'HH:mm:ss", outputFormat: String = "MMM dd, yyyy"): String {
    return try {
        val input = SimpleDateFormat(inputFormat, Locale.getDefault())
        val output = SimpleDateFormat(outputFormat, Locale.getDefault())
        val date = input.parse(this)
        date?.let { output.format(it) } ?: this
    } catch (e: Exception) {
        this
    }
}

fun Long.formatFileSize(): String {
    return when {
        this < 1024 -> "$this B"
        this < 1024 * 1024 -> "${this / 1024} KB"
        this < 1024 * 1024 * 1024 -> "${this / (1024 * 1024)} MB"
        else -> "${this / (1024 * 1024 * 1024)} GB"
    }
}

fun Long.formatDate(outputFormat: String = "MMM dd, yyyy"): String {
	return try {
		val output = SimpleDateFormat(outputFormat, Locale.getDefault())
		output.format(Date(this))
	} catch (e: Exception) {
		toString()
	}
}

fun applyTheme(theme: String?) {
    when (theme) {
        "Light" -> AppCompatDelegate.setDefaultNightMode(AppCompatDelegate.MODE_NIGHT_NO)
        "Dark" -> AppCompatDelegate.setDefaultNightMode(AppCompatDelegate.MODE_NIGHT_YES)
        else -> AppCompatDelegate.setDefaultNightMode(AppCompatDelegate.MODE_NIGHT_FOLLOW_SYSTEM)
    }
}
