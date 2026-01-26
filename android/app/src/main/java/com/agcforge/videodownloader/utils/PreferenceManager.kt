package com.agcforge.videodownloader.utils

import android.content.Context
import android.widget.ImageView
import android.widget.Toast
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.bumptech.glide.Glide
import com.bumptech.glide.load.resource.drawable.DrawableTransitionOptions
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import java.text.SimpleDateFormat
import java.util.*

// DataStore extension
val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "app_preferences")

// Preference Manager
class PreferenceManager(private val context: Context) {

    companion object {
        private val TOKEN_KEY = stringPreferencesKey("auth_token")
        private val USER_ID_KEY = stringPreferencesKey("user_id")
        private val USER_EMAIL_KEY = stringPreferencesKey("user_email")
        private val USER_NAME_KEY = stringPreferencesKey("user_name")
    }

    val authToken: Flow<String?> = context.dataStore.data.map { preferences ->
        preferences[TOKEN_KEY]
    }

    val userId: Flow<String?> = context.dataStore.data.map { preferences ->
        preferences[USER_ID_KEY]
    }

    suspend fun saveAuthToken(token: String) {
        context.dataStore.edit { preferences ->
            preferences[TOKEN_KEY] = token
        }
    }

    suspend fun saveUserInfo(userId: String, email: String, name: String) {
        context.dataStore.edit { preferences ->
            preferences[USER_ID_KEY] = userId
            preferences[USER_EMAIL_KEY] = email
            preferences[USER_NAME_KEY] = name
        }
    }

    suspend fun clearUserData() {
        context.dataStore.edit { preferences ->
            preferences.clear()
        }
    }
}

// Extension Functions
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
