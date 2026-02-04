package com.agcforge.videodownloader.helper

import android.app.Application
import android.util.Log
import com.google.android.gms.ads.MobileAds
import com.google.android.gms.ads.RequestConfiguration
import com.startapp.sdk.adsbase.StartAppSDK
import com.unity3d.ads.IUnityAdsInitializationListener
import com.unity3d.ads.UnityAds

object AdsInitializer {

    private const val TAG = "AdsInitializer"
    private var isInitialized = false

    /**
     * Initialize all ads SDKs
     */
    fun initialize(application: Application, onComplete: ((Boolean) -> Unit)? = null) {
        if (isInitialized) {
            Log.d(TAG, "Ads already initialized")
            onComplete?.invoke(true)
            return
        }

        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled in config")
            onComplete?.invoke(false)
            return
        }

        Log.d(TAG, "Initializing ads SDKs...")

        // Initialize Admob
        initializeAdmob(application)

        // Initialize Unity Ads
        initializeUnityAds(application)

        // Initialize Start.io
        initializeStartIo(application)

        isInitialized = true
        onComplete?.invoke(true)
    }

    /**
     * Initialize Admob
     */
    private fun initializeAdmob(application: Application) {
        try {
            Log.d(TAG, "Initializing Admob...")

            MobileAds.initialize(application) { initializationStatus ->
                val statusMap = initializationStatus.adapterStatusMap
                for (adapterClass in statusMap.keys) {
                    val status = statusMap[adapterClass]
                    Log.d(TAG, "Admob Adapter: $adapterClass, Status: ${status?.description}")
                }
                Log.d(TAG, "Admob initialized successfully")
            }

            // Configure request settings
            val requestConfiguration = RequestConfiguration.Builder()
                .setTagForChildDirectedTreatment(RequestConfiguration.TAG_FOR_CHILD_DIRECTED_TREATMENT_FALSE)
                .setTagForUnderAgeOfConsent(RequestConfiguration.TAG_FOR_UNDER_AGE_OF_CONSENT_FALSE)
                .build()

            MobileAds.setRequestConfiguration(requestConfiguration)

        } catch (e: Exception) {
            Log.e(TAG, "Admob initialization failed: ${e.message}")
        }
    }

    /**
     * Initialize Unity Ads
     */
    private fun initializeUnityAds(application: Application) {
        try {
            Log.d(TAG, "Initializing Unity Ads...")

            UnityAds.initialize(
                application,
                AdsConfig.UNITY_GAME_ID,
                AdsConfig.TEST_MODE,
                object : IUnityAdsInitializationListener {
                    override fun onInitializationComplete() {
                        Log.d(TAG, "Unity Ads initialized successfully")
                    }

                    override fun onInitializationFailed(
                        error: UnityAds.UnityAdsInitializationError,
                        message: String
                    ) {
                        Log.e(TAG, "Unity Ads initialization failed: $message")
                    }
                }
            )

        } catch (e: Exception) {
            Log.e(TAG, "Unity Ads initialization failed: ${e.message}")
        }
    }

    /**
     * Initialize Start.io
     */
    private fun initializeStartIo(application: Application) {
        try {
            Log.d(TAG, "Initializing Start.io...")

            AdsConfig.STARTIO_APP_ID?.let {
                StartAppSDK.init(
                    application,
                    it,
                    false // Set true for testing
                )
            }

            // Optional: Disable splash ads
//            StartAppSDK.setDisableSplash(true)

            // Optional: Set user consent for GDPR
            // StartAppSDK.setUserConsent(application, "pas", System.currentTimeMillis(), true)

            Log.d(TAG, "Start.io initialized successfully")

        } catch (e: Exception) {
            Log.e(TAG, "Start.io initialization failed: ${e.message}")
        }
    }

    /**
     * Check if ads are initialized
     */
    fun isInitialized(): Boolean {
        return isInitialized
    }
}