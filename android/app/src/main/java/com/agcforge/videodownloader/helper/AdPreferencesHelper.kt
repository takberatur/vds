package com.agcforge.videodownloader.helper

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit

class AdPreferencesHelper(context: Context) {
    companion object {
        private const val PREF_BANNER_AD_NAME = "banner_ad_preferences"
        private const val PREF_INTERSTITIAL_AD = "interstitial_ad_preferences"
        private const val PREF_NATIVE_AD = "native_ad_preferences"
        private const val KEY_LAST_SUCCESSFUL_BANNER_AD = "last_successful_banner_ad"
        private const val KEY_LAST_SUCCESSFUL_INTERSTITIAL_AD = "last_successful_interstitial_ad"
        private const val KEY_LAST_SUCCESSFUL_NATIVE_AD = "last_successful_native_ad"
        private const val KEY_BANNER_AD_ROTATION_INDEX = "banner_ad_rotation_index"
        private const val KEY_INTERSTITIAL_AD_ROTATION_INDEX = "interstitial_ad_rotation_index"
        private const val KEY_NATIVE_AD_ROTATION_INDEX = "native_ad_rotation_index"
        private const val KEY_LOAD_INTERSTITIAL_ATTEMPTS = "interstitial_load_attempts"
    }

    private val banner_ad_preferences: SharedPreferences =
        context.getSharedPreferences(PREF_BANNER_AD_NAME, Context.MODE_PRIVATE)
    private val interstitial_preferences: SharedPreferences =
        context.getSharedPreferences(PREF_INTERSTITIAL_AD, Context.MODE_PRIVATE)
    private val native_ad_preferences: SharedPreferences =
        context.getSharedPreferences(PREF_NATIVE_AD, Context.MODE_PRIVATE)

    // Banner Helper
    fun setLastBannerAdSuccessfulAd(adType: String) {
        banner_ad_preferences.edit { putString(KEY_LAST_SUCCESSFUL_BANNER_AD, adType) }
    }

    fun getLastBannerSuccessfulAd(): String =
        banner_ad_preferences.getString(KEY_LAST_SUCCESSFUL_BANNER_AD, "") ?: ""

    fun setBannerAdRotationIndex(index: Int) {
        banner_ad_preferences.edit { putInt(KEY_BANNER_AD_ROTATION_INDEX, index) }
    }

    fun getBannerAdRotationIndex(): Int =
        banner_ad_preferences.getInt(KEY_BANNER_AD_ROTATION_INDEX, 0)

    fun resetBannerAd() {
        banner_ad_preferences.edit { clear() }
    }

    // Interstitial Helper
    fun setLastInterstitialAdSuccessfulAd(adType: String) {
        interstitial_preferences.edit { putString(KEY_LAST_SUCCESSFUL_INTERSTITIAL_AD, adType) }
    }

    fun getLastInterstitialSuccessfulAd(): String =
        interstitial_preferences.getString(KEY_LAST_SUCCESSFUL_INTERSTITIAL_AD, "") ?: ""

    fun setInterstitialAdRotationIndex(index: Int) {
        interstitial_preferences.edit { putInt(KEY_INTERSTITIAL_AD_ROTATION_INDEX, index) }
    }

    fun getInterstitialAdRotationIndex(): Int =
        interstitial_preferences.getInt(KEY_INTERSTITIAL_AD_ROTATION_INDEX, 0)

    fun setLoadInterstitialAdAttempts(attempts: Int) {
        interstitial_preferences.edit { putInt(KEY_LOAD_INTERSTITIAL_ATTEMPTS, attempts) }
    }

    fun getLoadInterstitialAdAttempts(): Int =
        interstitial_preferences.getInt(KEY_LOAD_INTERSTITIAL_ATTEMPTS, 0)

    fun resetInterstitialAd() {
        interstitial_preferences.edit { clear() }
    }

    // Native Helper
    fun setLastSuccessfulNativeAd(adType: String) {
        native_ad_preferences.edit { putString(KEY_LAST_SUCCESSFUL_NATIVE_AD, adType) }
    }

    fun getLastSuccessfulNativeAd(): String =
        native_ad_preferences.getString(KEY_LAST_SUCCESSFUL_NATIVE_AD, "") ?: ""

    fun setNativeAdRotationIndex(index: Int) {
        native_ad_preferences.edit { putInt(KEY_NATIVE_AD_ROTATION_INDEX, index) }
    }

    fun getNativeAdRotationIndex(): Int =
        native_ad_preferences.getInt(KEY_NATIVE_AD_ROTATION_INDEX, 0)

    fun resetNativeAd() {
        native_ad_preferences.edit { clear() }
    }
}